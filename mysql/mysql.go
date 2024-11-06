package mysql

import (
	"encoding/json"
	"fmt"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var rds_map_gorm = make(map[string]*MysqlClient)
var rds_lock_gorm sync.RWMutex

type MysqlCfg struct {
	Host        string `json:"host"`
	Password    string `json:"password"`
	Dbname      string `json:"dbname"`
	User        string `json:"user"`
	Port        int    `json:"port"`
	Charset     string `json:"charset" default:"utf8mb4"`
	MaxIdleConn int    `json:"max_idle_conn" default:"10"`
	MaxOpenConn int    `json:"max_open_conn" default:"100"`
}
type MysqlClient struct {
	*gorm.DB
}

func GetInstanceGOrm(cfg *MysqlCfg) (*MysqlClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("logger or cfg is nil ")
	}
	if key, err := json.Marshal(cfg); err == nil {
		rds_lock_gorm.RLock()
		cli := rds_map_gorm[string(key)]
		rds_lock_gorm.RUnlock()
		if cli == nil {
			cli = &MysqlClient{}
			if db, err := cli.openGOrm(cfg); err != nil {
				return nil, err
			} else {
				cli.DB = db
				rds_lock_gorm.Lock()
				rds_map_gorm[string(key)] = cli
				rds_lock_gorm.Unlock()
			}
		}
		return cli, nil
	} else {
		return nil, err
	}
}
func (c *MysqlClient) openGOrm(cfg *MysqlCfg) (*gorm.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cfg is nil")
	}
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=%v&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Dbname,
		cfg.Charset,
	)
	eng, err := gorm.Open(mysql.Open(dsn)) /*, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})*/
	if err != nil {
		return nil, err
	}
	realDB, err := eng.DB()
	if err != nil {
		return nil, err
	}
	if cfg.MaxOpenConn > 0 {
		realDB.SetMaxOpenConns(cfg.MaxOpenConn)
	}
	if cfg.MaxIdleConn > 0 {
		realDB.SetMaxIdleConns(cfg.MaxIdleConn)
	}
	err = realDB.Ping()
	if err != nil {
		return nil, err
	}
	return eng, nil
}
