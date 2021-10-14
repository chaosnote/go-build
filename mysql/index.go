package mysql

/**
mysql 通用服務項目
// mysql 新手教學( 含:語法使用 )
// https://ithelp.ithome.com.tw/articles/10220925
// https://www.begtut.com/mysql/mysql-tutorial.html
*/

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // mysql-driver
)

func UnmarshalConfig(data []byte) (Config, error) {
	var r Config
	err := json.Unmarshal(data, &r)
	return r, err
}

type Config struct {
	Host     string `json:"host"`
	Schema   string `json:"schema"`
	Account  string `json:"account"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// New ...
func New(c Config) *sql.DB {
	// key word : mysql current connection info
	// sample 1. 當前連線資訊
	// mysql -u root -p
	// DB 登入後，可使用下列語法查詢
	// show status like '%connected';
	// show status like 'Conn%';
	// show processlist;
	// show variables like 'max_connections';
	// sample 2.
	// query
	// @see https://codertw.com/%E8%B3%87%E6%96%99%E5%BA%AB/10281/
	// 比較完整(中文)
	// @see https://chromium.googlesource.com/external/github.com/go-sql-driver/mysql/+/a48f79b55b5a2107793c84c3bbd445138dc7f0d5/README.md#examples
	// 設置規則
	p := fmt.Sprintf("%s:%s@%s/%s", c.Account, c.Password, c.Host, c.Name) // 對應 sql 連線格式

	db, e := sql.Open(c.Schema, p)
	//@see https://github.com/go-sql-driver/mysql#usage
	db.SetConnMaxLifetime(time.Minute * 1)
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)

	if e != nil {
		panic(e.Error())
	}
	// defer db.Close() // 此處不需關閉(關閉的話，後續會無法調用 )

	e = db.Ping()
	if e != nil {
		panic(e.Error())
	}

	return db
}

// // GoodsNO 商品代碼
// var GoodsNO []string
// // prepare 準備動作
// func prepare() {
// 	_, e := db.Exec(sp.AddDBStatus)
// 	if e != nil {
// 		panic(e.Error())
// 	}
// 	// 取得所有股票代碼( 可修改至 redis 或 伺服器開啟時取得一次<當前> )
// 	row, e := Get().Query("SELECT no FROM `" + table.NOAll + "` WHERE `exist`=1 ;")
// 	if e != nil {
// 		log.Get().Fatal(e.Error())
// 	}
// 	defer row.Close() // be careful
// 	var no string // 當參數與回傳值長度不相等時，會變成空字串
// 	for row.Next() {
// 		row.Scan(&no)
// 		GoodsNO = append(GoodsNO, no)
// 	}
// }

// // ToORM ORM 運用
// // https://ithelp.ithome.com.tw/articles/10220925
// // https://juejin.im/post/6844904090196000775
// func ToORM(rows *sql.Rows, dest interface{}) error {
// 	// 取得資料的每一列的名稱
// 	colNames, err := rows.Columns()
// 	if err != nil {
// 		return err
// 	}
// 	// 取得變數對象的值跟類型資訊
// 	v := reflect.ValueOf(dest)
// 	if v.Elem().Type().Kind() != reflect.Struct {
// 		return errors.New("give me  a struct")
// 	}
// 	// 宣告一個interface{}的slice
// 	scanDest := []interface{}{}
// 	// 建立一個string, interface{}的map
// 	addrByColName := map[string]interface{}{}
// 	for i := 0; i < v.Elem().NumField(); i++ {
// 		propertyName := v.Elem().Field(i)
// 		colName := v.Elem().Type().Field(i).Tag.Get("db")
// 		if colName == "" {
// 			if v.Elem().Field(i).CanInterface() == false {
// 				continue
// 			}
// 			colName = propertyName.Type().Name()
// 		}
// 		// Addr() 返回該屬性的記憶體位置的指針
// 		// Interface() 返回該屬性真正的值, 這裡還是存著位置
// 		addrByColName[colName] = propertyName.Addr().Interface()
// 	}
// 	// 把實際各成員屬性的位置, 給加到scan_dest中
// 	for _, colName := range colNames {
// 		scanDest = append(scanDest, addrByColName[colName])
// 	}
// 	// 執行Scan
// 	return rows.Scan(scanDest...)
// }
