package models

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

type Customers struct {
	CustomerId           int64                 `orm:"auto"`
	FullName             string                `orm:"column(full_name);size(255)"`
	ImagePath            string                `orm:"column(image_path);size(255)"`
	Email                string                `orm:"column(email);size(255);null"`
	PhoneNumber          string                `orm:"column(phone_number);size(255);null"`
	Location             string                `orm:"column(location);size(255);null"`
	IdentificationType   *Identification_types `orm:"rel(fk);column(identification_type_id);omitempty;null"`
	IdentificationNumber string                `orm:"column(identification_number);size(255);null"`
	CustomerCategory     *Customer_categories  `orm:"rel(fk);omitempty;null"`
	Nickname             string                `orm:"size(100);omitempty;null"`
	Dob                  time.Time             `orm:"column(dob);type(datetime)"`
	DateCreated          time.Time             `orm:"type(datetime)"`
	DateModified         time.Time             `orm:"type(datetime)"`
	CreatedBy            int
	ModifiedBy           int
	Active               int
	LastTxnDate          time.Time `orm:"type(datetime)"`
}

func init() {
	orm.RegisterModel(new(Customers))
}

// AddCustomers insert a new Customers into database and returns
// last inserted Id on success.
func AddCustomer(m *Customers) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
}

// GetCustomersById retrieves Customers by Id. Returns error if
// Id doesn't exist
func GetCustomerById(id int64) (v *Customers, err error) {
	o := orm.NewOrm()
	v = &Customers{CustomerId: id}
	qs := o.QueryTable(new(Customers))
	if err = qs.Filter("CustomerId", id).RelatedSel().One(v); err == nil {
		return v, nil
	}
	return nil, err
}

func GetCustomerByPhoneNumber(phoneNumber string) (v *Customers, err error) {
	o := orm.NewOrm()
	v = &Customers{PhoneNumber: phoneNumber}
	qs := o.QueryTable(new(Customers))
	if err = qs.Filter("PhoneNumber", phoneNumber).RelatedSel().One(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetItemsById retrieves Items by Id. Returns error if
// Id doesn't exist
func GetCustomerCount(query map[string]string, search map[string]string) (c int64, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(Customers))

	if len(search) > 0 {
		cond := orm.NewCondition()
		for k, v := range search {
			// rewrite dot-notation to Object__Attribute
			k = strings.Replace(k, ".", "__", -1)
			if strings.Contains(k, "isnull") {
				qs = qs.Filter(k, (v == "true" || v == "1"))
			} else {
				logs.Info("Adding or statement")
				cond = cond.Or(k+"__icontains", v)

				// qs = qs.Filter(k+"__icontains", v)

			}
		}
		logs.Info("Condition set ", qs)
		qs = qs.SetCond(cond)
	}

	logs.Info("Query is ", query)
	for k, v := range query {
		// rewrite dot-notation to Object__Attribute
		k = strings.Replace(k, ".", "__", -1)
		qs = qs.Filter(k, v)
	}
	if c, err = qs.RelatedSel().Count(); err == nil {
		return c, nil
	}
	return 0, err
}

// GetAllCustomers retrieves all Customers matches certain condition. Returns empty list if
// no records exist
func GetAllCustomers(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64, search map[string]string) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(Customers))

	if len(search) > 0 {
		cond := orm.NewCondition()
		for k, v := range search {
			// rewrite dot-notation to Object__Attribute
			k = strings.Replace(k, ".", "__", -1)
			if strings.Contains(k, "isnull") {
				qs = qs.Filter(k, (v == "true" || v == "1"))
			} else {
				logs.Info("Adding or statement")
				cond = cond.Or(k+"__icontains", v)

				// qs = qs.Filter(k+"__icontains", v)

			}
		}
		logs.Info("Condition set ", qs)
		qs = qs.SetCond(cond)
	}

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

	var l []Customers
	qs = qs.OrderBy(sortFields...).RelatedSel()

	// Without Limit
	if _, err = qs.All(&l, fields...); err == nil {
		if len(fields) == 0 {
			for _, v := range l {
				_, err = o.LoadRelated(&v, "EmergencyContacts")
				if err != nil {
					logs.Error("Error loading related EmergencyContacts: ", err)
				} else {
					logs.Info("EmergencyContacts loaded successfully")
					// fmt.Printf("Emergency contacts of customers: %T\n", v)
					// fmt.Printf("Emergency contacts of customers: %v\n", v.EmergencyContacts)
					// for _, ec := range v.EmergencyContacts {
					// 	fmt.Printf("Emergency contact: %v\n", ec)
					// }
				}
				_, err = o.LoadRelated(&v, "Guarantors")
				if err != nil {
					logs.Error("Error loading related Guarantors: ", err)
				}
				ml = append(ml, v)
			}
		} else {
			// trim unused fields
			for _, v := range l {
				_, err = o.LoadRelated(&v, "EmergencyContacts")
				if err != nil {
					logs.Error("Error loading related EmergencyContacts: ", err)
				} else {
					logs.Info("EmergencyContacts loaded successfully")
					// fmt.Printf("Emergency contacts of customers: %T\n", v)
					// fmt.Printf("Emergency contacts of customers: %v\n", v.EmergencyContacts)
					// for _, ec := range v.EmergencyContacts {
					// 	fmt.Printf("Emergency contact: %v\n", ec)
					// }
				}
				_, err = o.LoadRelated(&v, "Guarantors")
				if err != nil {
					logs.Error("Error loading related Guarantors: ", err)
				}
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

	// With limit
	// if _, err = qs.Limit(limit, offset).All(&l, fields...); err == nil {
	// 	if len(fields) == 0 {
	// 		for _, v := range l {
	// 			ml = append(ml, v)
	// 		}
	// 	} else {
	// 		// trim unused fields
	// 		for _, v := range l {
	// 			m := make(map[string]interface{})
	// 			val := reflect.ValueOf(v)
	// 			for _, fname := range fields {
	// 				m[fname] = val.FieldByName(fname).Interface()
	// 			}
	// 			ml = append(ml, m)
	// 		}
	// 	}
	// 	return ml, nil
	// }
	return nil, err
}

// UpdateCustomers updates Customers by Id and returns error if
// the record to be updated doesn't exist
func UpdateCustomerById(m *Customers) (err error) {
	o := orm.NewOrm()
	v := Customers{CustomerId: m.CustomerId}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			fmt.Println("Number of records updated in database:", num)
		}
	}
	return
}

// DeleteCustomers deletes Customers by Id and returns error if
// the record to be deleted doesn't exist
func DeleteCustomer(id int64) (err error) {
	o := orm.NewOrm()
	v := Customers{CustomerId: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&Customers{CustomerId: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}
