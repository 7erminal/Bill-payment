package models

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Bil_ins_transactions struct {
	BilInsTransactionId    int64             `orm:"auto"`
	BilTransactionId       *Bil_transactions `orm:"rel(fk);column(bil_transaction_id)"`
	Amount                 float64
	Biller                 *Billers  `orm:"rel(fk);column(biller_id)"`
	SenderAccountNumber    string    `orm:"size(255)"`
	RecipientAccountNumber string    `orm:"size(255)"`
	Network                string    `orm:"size(150)"`
	Request                string    `orm:"size(255)"`
	Response               string    `orm:"size(255)"`
	DateCreated            time.Time `orm:"type(datetime)"`
	DateModified           time.Time `orm:"type(datetime)"`
	CreatedBy              int
	ModifiedBy             int
	Active                 int
}

func init() {
	orm.RegisterModel(new(Bil_ins_transactions))
}

// AddBil_ins_transactions insert a new Bil_ins_transactions into database and returns
// last inserted Id on success.
func AddBil_ins_transactions(m *Bil_ins_transactions) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
}

// GetBil_ins_transactionsById retrieves Bil_ins_transactions by Id. Returns error if
// Id doesn't exist
func GetBil_ins_transactionsById(id int64) (v *Bil_ins_transactions, err error) {
	o := orm.NewOrm()
	v = &Bil_ins_transactions{BilInsTransactionId: id}
	if err = o.QueryTable(new(Bil_ins_transactions)).Filter("BilInsTransactionId", id).RelatedSel().One(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetAllBil_ins_transactions retrieves all Bil_ins_transactions matches certain condition. Returns empty list if
// no records exist
func GetAllBil_ins_transactions(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(Bil_ins_transactions))
	// query k=v
	for k, v := range query {
		// rewrite dot-notation to Object__Attribute
		k = strings.Replace(k, ".", "__", -1)
		qs = qs.Filter(k, v)
	}
	// order by:
	var sortFields []string
	if len(sortby) != 0 {
		if len(sortby) == len(order) {
			// 1) for each sort field, there is an associated order
			for i, v := range sortby {
				orderby := ""
				if order[i] == "desc" {
					orderby = "-" + v
				} else if order[i] == "asc" {
					orderby = v
				} else {
					return nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
			qs = qs.OrderBy(sortFields...)
		} else if len(sortby) != len(order) && len(order) == 1 {
			// 2) there is exactly one order, all the sorted fields will be sorted by this order
			for _, v := range sortby {
				orderby := ""
				if order[0] == "desc" {
					orderby = "-" + v
				} else if order[0] == "asc" {
					orderby = v
				} else {
					return nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
		} else if len(sortby) != len(order) && len(order) != 1 {
			return nil, errors.New("Error: 'sortby', 'order' sizes mismatch or 'order' size is not 1")
		}
	} else {
		if len(order) != 0 {
			return nil, errors.New("Error: unused 'order' fields")
		}
	}

	var l []Bil_ins_transactions
	qs = qs.OrderBy(sortFields...).RelatedSel()
	if _, err = qs.Limit(limit, offset).All(&l, fields...); err == nil {
		if len(fields) == 0 {
			for _, v := range l {
				ml = append(ml, v)
			}
		} else {
			// trim unused fields
			for _, v := range l {
				m := make(map[string]interface{})
				val := reflect.ValueOf(v)
				for _, fname := range fields {
					m[fname] = val.FieldByName(fname).Interface()
				}
				ml = append(ml, m)
			}
		}
		return ml, nil
	}
	return nil, err
}

// UpdateBil_ins_transactions updates Bil_ins_transactions by Id and returns error if
// the record to be updated doesn't exist
func UpdateBil_ins_transactionsById(m *Bil_ins_transactions) (err error) {
	o := orm.NewOrm()
	v := Bil_ins_transactions{BilInsTransactionId: m.BilInsTransactionId}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			fmt.Println("Number of records updated in database:", num)
		}
	}
	return
}

// DeleteBil_ins_transactions deletes Bil_ins_transactions by Id and returns error if
// the record to be deleted doesn't exist
func DeleteBil_ins_transactions(id int64) (err error) {
	o := orm.NewOrm()
	v := Bil_ins_transactions{BilInsTransactionId: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&Bil_ins_transactions{BilInsTransactionId: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}
