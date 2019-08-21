package mysqltoproto

import "testing"

func TestGenerateProto(t *testing.T) {
	object, err := NewMysqlToProto("D:/Github/mysqltoproto/config.cfg")

	if nil != err {
		t.Error("NewMysqlToProto Err", err)
	}

	err = object.GenerateProto()

	if nil != err {
		t.Error("GenerateProto Err, ", err)
	}
}
