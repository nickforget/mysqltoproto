package mysqltoproto

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/nickforget/dboperate"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"github.com/Unknwon/goconfig"
)

/*
	CREATE VIEW mytable AS SELECT
		table_name tablename,
		column_name columnname,
		column_comment columncomment,
		data_type datatype,
		table_schema tableschema
	FROM
		information_schema. COLUMNS
*/

var DataTypeTrans map[string]string = map[string]string{
	"int" : "int32",
	"blob" : "bytes",
	"text" : "string",
	"char" : "string",
	"bigint" : "int64",
	"longblob" : "bytes",
	"decimal" : "double",
	"varchar" : "string",
	"timestamp" : "string",
}

type Proto struct {
	DataType      string
	ColumnName    string
	ColumnComment string
}

type MysqlToProto struct {
	DestPath string
	DestFileName string
	DBConnStr string
	TableSchema string
	TableName []string
}

func NewMysqlToProto(cfgpath string) (*MysqlToProto, error){
	cfg, err := goconfig.LoadConfigFile(cfgpath)
	if err != nil{
		return nil, err
	}

	destpath, err := cfg.GetValue("config", "destpath")

	if err != nil{
		return nil, err
	}

	dbconnstr, err := cfg.GetValue("config", "dbconnstr")

	if err != nil{
		return nil, err
	}

	tablename, err := cfg.GetValue("config", "tablename")

	if err != nil{
		return nil, err
	}

	tableschema, err := cfg.GetValue("config", "tableschema")

	if err != nil{
		return nil, err
	}

	destfilename, err := cfg.GetValue("config", "destfilename")

	if err != nil{
		return nil, err
	}

	return &MysqlToProto{
		DestPath : destpath,
		DestFileName : destfilename,
		DBConnStr : dbconnstr,
		TableSchema : tableschema,
		TableName : strings.Fields(tablename),
	}, nil
}

const MAXBYTESIZE = 100

func (this *MysqlToProto) GenerateProto() error{
	// 读取数据
	data, err := this.ReadDB()

	if nil != err{
		return err
	}

	// 生成文件
	err = this.WriteFile(data)

	if nil != err {
		return err
	}

	return nil
}

func (this *MysqlToProto) ReadDB() (map[string][]Proto, error){
	// 打开数据库
	db := dboperate.NewDBOperate("mysql", this.DBConnStr)

	if nil == db{
		return nil, errors.New("NewDBOperate Err")
	}

	// 连接数据库
	err := db.ConnDB()

	if nil != err {
		return nil, errors.New("Conn DB Err")
	}

	// 查询数据库表和字段
	records, err := db.Query("mytable",[]string{},"", &MyTable{TableSchema:proto.String(this.TableSchema)})

	if nil != err{
		return nil, errors.New("Query DB Err")
	}

	// 定义协议数据
	protodata := make(map[string]([]Proto), 0)

	// 写入协议数据
	for _, record := range records{
		comment := *record.(*MyTable).ColumnComment

		// 去除空格
		comment = strings.Replace(comment, " ", "", -1)

		// 去除回车
		comment = strings.Replace(comment, "\r\n", "", -1)

		// 转成小写
		comment = strings.ToLower(comment)

		// 去掉注释
		if len(comment) > MAXBYTESIZE {
			comment = " "
		}

		data := Proto{
			DataType : DataTypeTrans[*record.(*MyTable).DataType],
			ColumnName : strings.ToLower(*record.(*MyTable).ColumnName),
			ColumnComment : comment,
		}
		protodata[*record.(*MyTable).TableName] = append(protodata[*record.(*MyTable).TableName], data)
	}


	return protodata, nil
}

func (this *MysqlToProto) WriteFile(protodata map[string][]Proto) (error){
	// 生成proto
	data, err := ioutil.ReadFile("proto.tpl")

	if err != nil {
		return errors.New("Read Proto.tpl Err")
	}

	// 设置模板函数
	funcmap := template.FuncMap{
		"add" : func(num int)int{
			return num + 1
		},
	}

	prototp, err := template.New("prototp").Funcs(funcmap).Parse(string(data))
	if err != nil{
		return errors.New("New Template Err")
	}

	// 创建文件
	fp, err := os.Create(this.DestPath + this.DestFileName)

	if err != nil{
		fmt.Println("Create Err, Err : ", err.Error())
		return errors.New("Create Dest File Err")
	}

	// 关闭文件
	defer fp.Close()

	// 执行模板
	err = prototp.Execute(fp, protodata)

	if nil != err {
		return errors.New("Execute Template Err")
	}

	return nil
}

