package clause_test

import (
	"geeorm/clause"
	"reflect"
	"testing"
)

func testSelect(t *testing.T) {
	c := clause.Clause{}
	c.Set(clause.LIMIT, 3)
	c.Set(clause.SELECT, "User", []string{"*"})
	c.Set(clause.WHERE, "Name = ?", "Tom")
	c.Set(clause.ORDERBY, "Age DESC")
	sql, vars := c.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	t.Log(sql, vars)
	if sql != "SELECT * FROM User WHERE Name = ? ORDER BY Age DESC LIMIT ?" {
		t.Fatal("failed to build SQL")
	}
	if !reflect.DeepEqual(vars, []interface{}{"Tom", 3}) {
		t.Fatal("failed to build SQL vars")
	}
}

func TestClause_Build(t *testing.T) {
	t.Run("SELECT", testSelect)
}
