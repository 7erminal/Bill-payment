package models

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Application_property struct {
	ApplicationPropertyId int64     `orm:"auto"`
	PropertyCode          string    `orm:"size(80)"`
	PropertyValue         string    `orm:"size(255)"`
	DateCreated           time.Time `orm:"type(datetime)"`
	DateModified          time.Time `orm:"type(datetime)"`
	CreatedBy             int
	ModifiedBy            int
	Active                int
}

func (t *Application_property) TableName() string {
	return "application_properties"
}

func init() {
	orm.RegisterModel(new(Application_property))
}

// AddApplication_property insert a new Application_property into database and returns
// last inserted Id on success.
func AddApplication_property(m *Application_property) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
}

// GetApplication_propertyById retrieves Application_property by Id. Returns error if
// Id doesn't exist
func GetApplication_propertyById(id int64) (v *Application_property, err error) {
	o := orm.NewOrm()
	v = &Application_property{ApplicationPropertyId: id}
	if err = o.QueryTable(new(Application_property)).Filter("ApplicationPropertyId", id).RelatedSel().One(v); err == nil {
		return v, nil
	}
	return nil, err
}

func GetApplication_propertyByCode(code string) (v *Application_property, err error) {
	o := orm.NewOrm()
	v = &Application_property{PropertyCode: code}
	if err = o.QueryTable(new(Application_property)).Filter("PropertyCode", code).RelatedSel().One(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetAllApplication_property retrieves all Application_property matches certain condition. Returns empty list if
// no records exist
func GetAllApplication_property(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(Application_property))
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

	var l []Application_property
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

// UpdateApplication_property updates Application_property by Id and returns error if
// the record to be updated doesn't exist
func UpdateApplication_propertyById(m *Application_property) (err error) {
	o := orm.NewOrm()
	v := Application_property{ApplicationPropertyId: m.ApplicationPropertyId}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			fmt.Println("Number of records updated in database:", num)
		}
	}
	return
}

// DeleteApplication_property deletes Application_property by Id and returns error if
// the record to be deleted doesn't exist
func DeleteApplication_property(id int64) (err error) {
	o := orm.NewOrm()
	v := Application_property{ApplicationPropertyId: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&Application_property{ApplicationPropertyId: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}
