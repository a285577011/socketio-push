package models

import (
	"errors"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

var (
	O orm.Ormer
)
var DeviceM *DeviceModel
func init() {

	DBInitNew().InitDatabase()
	orm.RegisterModel(new(Device))
	O = orm.NewOrm()
	O.Using("default")
	DeviceM=&DeviceModel{}
}
type DeviceModel struct {

}
type Device struct {
	Id       string `orm:"column(id);pk" description:"id"`
	Name string `orm:"column(name)" description:"name"`
	Imie string `orm:"column(imie)" description:"imie"`
	RegId   string `orm:"column(reg_id)" description:"reg_id"`
	Ip      string    `orm:"column(ip)" description:"ip"`
	Phone  string `orm:"column(phone)" description:"phone"`
	TaskTime    uint64 `orm:"column(task_time)" description:"task_time"`
	TaskStatus    int `orm:"column(task_status)" description:"task_status"`
	Status    int `orm:"column(status)" description:"status"`
	Ctime    string `orm:"column(c_time)" description:"c_time"`
}


func (this *DeviceModel)GetByImie(imie string) (d *Device, err error) {
	var device Device
	err = O.QueryTable("device").Filter("Imie", imie).One(&device)
	if err != nil {
		return nil, errors.New("device not exists")
	}

	return &device, nil

}