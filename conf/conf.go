package conf

import (
	"DBTools/utils"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/magiconair/properties"
	"os"
	"reflect"
)

var (
	DbS DbConf
)

type DbConf struct {
	ConfigFile string
	Username   string
	Password   string
	Hostname   string
	Port       int
	Version    string
	IsDBA      bool
	ChangePWD  bool
	Db         *sqlx.DB
	DataBase   string
}

func (D *DbConf) CreateConnection() error {
	if exists, err := D.dataBaseExists("irds_irdsdb"); err == nil && exists {
		D.DataBase = "irds_irdsdb"
	} else if exists, err := D.dataBaseExists("isupm_upmdb"); err == nil && exists {
		D.DataBase = "isupm_upmdb"
	} else {
		fmt.Println(utils.ColorPrint(-1, "Not Support"))
		os.Exit(0)
	}
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", D.Username, D.Password, D.Hostname, D.Port, D.DataBase)
	db, err := sqlx.Open("postgres", connectionString)
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	D.Db = db
	D.GetVersion()
	D.IsDba()
	D.ShowDbDetail()
	return nil
}
func (D *DbConf) dataBaseExists(dbName string) (bool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/?sslmode=disable", D.Username, D.Password, D.Hostname, D.Port)
	db, err := sql.Open("postgres", dsn)
	defer db.Close()
	if err != nil {
		return false, err
	}
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
func (D *DbConf) GetVersion() {
	Version := ""
	err := D.Db.QueryRow("SELECT version()").Scan(&Version)
	if err != nil {
		return
	}
	D.Version = Version
}
func (D *DbConf) IsDba() {
	Dba := false
	isDba := ""
	err := D.Db.QueryRow("SHOW is_superuser").Scan(&isDba)
	if err != nil {
		fmt.Println(utils.ColorPrint(-1, "Check DBA Failed"))
		return
	}
	if isDba == "on" {
		Dba = true
	}
	D.IsDBA = Dba
}
func (D DbConf) ShowDbDetail() {
	var typeInfo = reflect.TypeOf(D)
	var valInfo = reflect.ValueOf(D)
	num := typeInfo.NumField()
	for i := 0; i < num; i++ {
		key := typeInfo.Field(i).Name
		val := valInfo.Field(i).Interface()
		if key == "Db" {
			continue
		}
		fmt.Println(utils.ColorPrint(0, fmt.Sprintf("%v ==> %v", key, val)))
	}
}
func (D *DbConf) ParseConfigFile() {
	p := properties.MustLoadFile(D.ConfigFile, properties.UTF8)

	// 获取单个属性值
	port := p.GetInt("rdbms.1.port", 7092)
	password, exist := p.Get("rdbms.1.password")
	if exist {
		D.Password, _ = utils.DecryptData(password)
	}
	username := p.GetString("rdbms.1.username", "postgres")
	hostname := p.GetString("rdbms.1.@ip", "127.0.0.1")
	D.Username = username
	D.Port = port
	D.Hostname = hostname
}
