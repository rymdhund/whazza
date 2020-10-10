package persist

import (
	"testing"

	"github.com/rymdhund/whazza/internal/chk"
)

func TestCreateCheck(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	err = db.Init()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	agent := AgentModel{1, "agent"}
	if err = tx.SaveAgent(agent.Name, ""); err != nil {
		t.Fatal(err)
	}
	check, err := chk.New("http-up", "ns", 3, []byte(`{"host":"example.com"}`))
	if err != nil {
		t.Fatal(err)
	}
	cm, err := tx.RegisterCheck(agent, check)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}

	checkModel, err := db.GetCheckById(cm.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !chk.Equal(checkModel.Check, cm.Check) {
		t.Errorf("%v != %v", checkModel.Check, cm.Check)
	}
}
