// Code generated by gorm-gen. DO NOT EDIT.
// generated at 03 Nov 20 15:30 CST

package gorm

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

type (
	Association = gorm.Association
	DB          = gorm.DB
	Migrator    = gorm.Migrator
	Plugin      = gorm.Plugin
	Session     = gorm.Session
	Statement   = gorm.Statement
)

type Interface interface {
	AddError(err error) (err1 error)
	Assign(vals ...interface{}) (tx Interface)
	Association(field string) (association *Association)
	Attrs(vals ...interface{}) (tx Interface)
	AutoMigrate(vals ...interface{}) (err error)
	Begin(vals ...*sql.TxOptions) (tx Interface)
	BindVarTo(iface Writer, statement *Statement, iface1 interface{})
	Clauses(vals ...Expression) (tx Interface)
	Commit() (tx Interface)
	Count(ctx context.Context, ptr *int64) (db *DB)
	Create(ctx context.Context, iface interface{}) (db *DB)
	DB() (db *sql.DB, err error)
	DataTypeOf(field *Field) (ret string)
	Debug() (tx Interface)
	DefaultValueOf(ctx context.Context, field *Field) (iface Expression)
	Delete(ctx context.Context, iface interface{}, vals ...interface{}) (db *DB)
	Distinct(vals ...interface{}) (tx Interface)
	Exec(ctx context.Context, field string, vals ...interface{}) (db *DB)
	Explain(field string, vals ...interface{}) (ret string)
	Find(ctx context.Context, iface interface{}, vals ...interface{}) (db *DB)
	FindInBatches(ctx context.Context, iface interface{}, i int, fn func(*DB, int) error) (db *DB)
	First(ctx context.Context, iface interface{}, vals ...interface{}) (db *DB)
	FirstOrCreate(ctx context.Context, iface interface{}, vals ...interface{}) (db *DB)
	FirstOrInit(ctx context.Context, iface interface{}, vals ...interface{}) (db *DB)
	Get(ctx context.Context, field string) (iface interface{}, b bool)
	Group(field string) (tx Interface)
	Having(iface interface{}, vals ...interface{}) (tx Interface)
	InstanceGet(ctx context.Context, field string) (iface interface{}, b bool)
	InstanceSet(field string, iface interface{}) (tx Interface)
	Joins(field string, vals ...interface{}) (tx Interface)
	Last(ctx context.Context, iface interface{}, vals ...interface{}) (db *DB)
	Limit(i int) (tx Interface)
	Migrator(ctx context.Context) (iface Migrator)
	Model(iface interface{}) (tx Interface)
	Name() (ret string)
	Not(iface interface{}, vals ...interface{}) (tx Interface)
	Offset(i int) (tx Interface)
	Omit(vals ...string) (tx Interface)
	Or(iface interface{}, vals ...interface{}) (tx Interface)
	Order(iface interface{}) (tx Interface)
	Pluck(ctx context.Context, field string, iface interface{}) (db *DB)
	Preload(field string, vals ...interface{}) (tx Interface)
	QuoteTo(iface Writer, val string)
	Raw(field string, vals ...interface{}) (tx Interface)
	Rollback() (tx Interface)
	RollbackTo(field string) (tx Interface)
	Row(ctx context.Context) (row *sql.Row)
	Rows(ctx context.Context) (rows *sql.Rows, err error)
	Save(ctx context.Context, iface interface{}) (db *DB)
	SavePoint(field string) (tx Interface)
	Scan(ctx context.Context, iface interface{}) (db *DB)
	ScanRows(ctx context.Context, rows *sql.Rows, iface interface{}) (err error)
	Scopes(funcs ...func(*DB) *DB) (tx Interface)
	Select(iface interface{}, vals ...interface{}) (tx Interface)
	Session(session *Session) (tx Interface)
	Set(field string, iface interface{}) (tx Interface)
	SetupJoinTable(iface interface{}, val string, iface1 interface{}) (err error)
	Table(field string, vals ...interface{}) (tx Interface)
	Take(ctx context.Context, iface interface{}, vals ...interface{}) (db *DB)
	Transaction(fn func(*DB) error, vals ...*sql.TxOptions) (err error)
	Unscoped() (tx Interface)
	Update(ctx context.Context, field string, iface interface{}) (db *DB)
	UpdateColumn(ctx context.Context, field string, iface interface{}) (db *DB)
	UpdateColumns(ctx context.Context, iface interface{}) (db *DB)
	Updates(ctx context.Context, iface interface{}) (db *DB)
	Use(iface Plugin) (err error)
	Where(iface interface{}, vals ...interface{}) (tx Interface)
	WithContext(iface context.Context) (tx Interface)
}

func (c *Client) AddError(err error) (err1 error) {

	return c.load().db.AddError(err)

}

func (c *Client) Assign(vals ...interface{}) (tx Interface) {
	client := c.load()
	client.db = client.db.Assign(vals...)

	return client
}

func (c *Client) Association(field string) (association *Association) {

	return c.load().db.Association(field)

}

func (c *Client) Attrs(vals ...interface{}) (tx Interface) {
	client := c.load()
	client.db = client.db.Attrs(vals...)

	return client
}

func (c *Client) AutoMigrate(vals ...interface{}) (err error) {

	return c.load().db.AutoMigrate(vals...)

}

func (c *Client) Begin(vals ...*sql.TxOptions) (tx Interface) {
	client := c.load()
	client.db = client.db.Begin(vals...)

	return client
}

func (c *Client) BindVarTo(iface Writer, statement *Statement, iface1 interface{}) {

	c.load().db.BindVarTo(iface, statement, iface1)

}

func (c *Client) Clauses(vals ...Expression) (tx Interface) {
	client := c.load()
	client.db = client.db.Clauses(vals...)

	return client
}

func (c *Client) Commit() (tx Interface) {
	client := c.load()
	client.db = client.db.Commit()

	return client
}

func (c *Client) Count(ctx context.Context, ptr *int64) (db *DB) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	db = client.db.Count(ptr)
	return
}

func (c *Client) Create(ctx context.Context, iface interface{}) (db *DB) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	db = client.db.Create(iface)
	return
}

func (c *Client) DB() (db *sql.DB, err error) {

	return c.load().db.DB()

}

func (c *Client) DataTypeOf(field *Field) (ret string) {

	return c.load().db.DataTypeOf(field)

}

func (c *Client) Debug() (tx Interface) {
	client := c.load()
	client.db = client.db.Debug()

	return client
}

func (c *Client) DefaultValueOf(ctx context.Context, field *Field) (iface Expression) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	iface = client.db.DefaultValueOf(field)
	return
}

func (c *Client) Delete(ctx context.Context, iface interface{}, vals ...interface{}) (db *DB) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	db = client.db.Delete(iface, vals...)
	return
}

func (c *Client) Distinct(vals ...interface{}) (tx Interface) {
	client := c.load()
	client.db = client.db.Distinct(vals...)

	return client
}

func (c *Client) Exec(ctx context.Context, field string, vals ...interface{}) (db *DB) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	db = client.db.Exec(field, vals...)
	return
}

func (c *Client) Explain(field string, vals ...interface{}) (ret string) {

	return c.load().db.Explain(field, vals...)

}

func (c *Client) Find(ctx context.Context, iface interface{}, vals ...interface{}) (db *DB) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	db = client.db.Find(iface, vals...)
	return
}

func (c *Client) FindInBatches(ctx context.Context, iface interface{}, i int, fn func(*DB, int) error) (db *DB) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	db = client.db.FindInBatches(iface, i, fn)
	return
}

func (c *Client) First(ctx context.Context, iface interface{}, vals ...interface{}) (db *DB) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	db = client.db.First(iface, vals...)
	return
}

func (c *Client) FirstOrCreate(ctx context.Context, iface interface{}, vals ...interface{}) (db *DB) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	db = client.db.FirstOrCreate(iface, vals...)
	return
}

func (c *Client) FirstOrInit(ctx context.Context, iface interface{}, vals ...interface{}) (db *DB) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	db = client.db.FirstOrInit(iface, vals...)
	return
}

func (c *Client) Get(ctx context.Context, field string) (iface interface{}, b bool) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	iface, b = client.db.Get(field)
	return
}

func (c *Client) Group(field string) (tx Interface) {
	client := c.load()
	client.db = client.db.Group(field)

	return client
}

func (c *Client) Having(iface interface{}, vals ...interface{}) (tx Interface) {
	client := c.load()
	client.db = client.db.Having(iface, vals...)

	return client
}

func (c *Client) InstanceGet(ctx context.Context, field string) (iface interface{}, b bool) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	iface, b = client.db.InstanceGet(field)
	return
}

func (c *Client) InstanceSet(field string, iface interface{}) (tx Interface) {
	client := c.load()
	client.db = client.db.InstanceSet(field, iface)

	return client
}

func (c *Client) Joins(field string, vals ...interface{}) (tx Interface) {
	client := c.load()
	client.db = client.db.Joins(field, vals...)

	return client
}

func (c *Client) Last(ctx context.Context, iface interface{}, vals ...interface{}) (db *DB) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	db = client.db.Last(iface, vals...)
	return
}

func (c *Client) Limit(i int) (tx Interface) {
	client := c.load()
	client.db = client.db.Limit(i)

	return client
}

func (c *Client) Migrator(ctx context.Context) (iface Migrator) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	iface = client.db.Migrator()
	return
}

func (c *Client) Model(iface interface{}) (tx Interface) {
	client := c.load()
	client.db = client.db.Model(iface)

	return client
}

func (c *Client) Name() (ret string) {

	return c.load().db.Name()

}

func (c *Client) Not(iface interface{}, vals ...interface{}) (tx Interface) {
	client := c.load()
	client.db = client.db.Not(iface, vals...)

	return client
}

func (c *Client) Offset(i int) (tx Interface) {
	client := c.load()
	client.db = client.db.Offset(i)

	return client
}

func (c *Client) Omit(vals ...string) (tx Interface) {
	client := c.load()
	client.db = client.db.Omit(vals...)

	return client
}

func (c *Client) Or(iface interface{}, vals ...interface{}) (tx Interface) {
	client := c.load()
	client.db = client.db.Or(iface, vals...)

	return client
}

func (c *Client) Order(iface interface{}) (tx Interface) {
	client := c.load()
	client.db = client.db.Order(iface)

	return client
}

func (c *Client) Pluck(ctx context.Context, field string, iface interface{}) (db *DB) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	db = client.db.Pluck(field, iface)
	return
}

func (c *Client) Preload(field string, vals ...interface{}) (tx Interface) {
	client := c.load()
	client.db = client.db.Preload(field, vals...)

	return client
}

func (c *Client) QuoteTo(iface Writer, val string) {

	c.load().db.QuoteTo(iface, val)

}

func (c *Client) Raw(field string, vals ...interface{}) (tx Interface) {
	client := c.load()
	client.db = client.db.Raw(field, vals...)

	return client
}

func (c *Client) Rollback() (tx Interface) {
	client := c.load()
	client.db = client.db.Rollback()

	return client
}

func (c *Client) RollbackTo(field string) (tx Interface) {
	client := c.load()
	client.db = client.db.RollbackTo(field)

	return client
}

func (c *Client) Row(ctx context.Context) (row *sql.Row) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	row = client.db.Row()
	return
}

func (c *Client) Rows(ctx context.Context) (rows *sql.Rows, err error) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	rows, err = client.db.Rows()
	return
}

func (c *Client) Save(ctx context.Context, iface interface{}) (db *DB) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	db = client.db.Save(iface)
	return
}

func (c *Client) SavePoint(field string) (tx Interface) {
	client := c.load()
	client.db = client.db.SavePoint(field)

	return client
}

func (c *Client) Scan(ctx context.Context, iface interface{}) (db *DB) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	db = client.db.Scan(iface)
	return
}

func (c *Client) ScanRows(ctx context.Context, rows *sql.Rows, iface interface{}) (err error) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	err = client.db.ScanRows(rows, iface)
	return
}

func (c *Client) Scopes(funcs ...func(*DB) *DB) (tx Interface) {
	client := c.load()
	client.db = client.db.Scopes(funcs...)

	return client
}

func (c *Client) Select(iface interface{}, vals ...interface{}) (tx Interface) {
	client := c.load()
	client.db = client.db.Select(iface, vals...)

	return client
}

func (c *Client) Session(session *Session) (tx Interface) {
	client := c.load()
	client.db = client.db.Session(session)

	return client
}

func (c *Client) Set(field string, iface interface{}) (tx Interface) {
	client := c.load()
	client.db = client.db.Set(field, iface)

	return client
}

func (c *Client) SetupJoinTable(iface interface{}, val string, iface1 interface{}) (err error) {

	return c.load().db.SetupJoinTable(iface, val, iface1)

}

func (c *Client) Table(field string, vals ...interface{}) (tx Interface) {
	client := c.load()
	client.db = client.db.Table(field, vals...)

	return client
}

func (c *Client) Take(ctx context.Context, iface interface{}, vals ...interface{}) (db *DB) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	db = client.db.Take(iface, vals...)
	return
}

func (c *Client) Transaction(fn func(*DB) error, vals ...*sql.TxOptions) (err error) {

	return c.load().db.Transaction(fn, vals...)

}

func (c *Client) Unscoped() (tx Interface) {
	client := c.load()
	client.db = client.db.Unscoped()

	return client
}

func (c *Client) Update(ctx context.Context, field string, iface interface{}) (db *DB) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	db = client.db.Update(field, iface)
	return
}

func (c *Client) UpdateColumn(ctx context.Context, field string, iface interface{}) (db *DB) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	db = client.db.UpdateColumn(field, iface)
	return
}

func (c *Client) UpdateColumns(ctx context.Context, iface interface{}) (db *DB) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	db = client.db.UpdateColumns(iface)
	return
}

func (c *Client) Updates(ctx context.Context, iface interface{}) (db *DB) {
	client := c.load()
	if ctx != nil {
		client.db = client.db.WithContext(ctx)
	}

	db = client.db.Updates(iface)
	return
}

func (c *Client) Use(iface Plugin) (err error) {

	return c.load().db.Use(iface)

}

func (c *Client) Where(iface interface{}, vals ...interface{}) (tx Interface) {
	client := c.load()
	client.db = client.db.Where(iface, vals...)

	return client
}

func (c *Client) WithContext(iface context.Context) (tx Interface) {
	client := c.load()
	client.db = client.db.WithContext(iface)

	return client
}
