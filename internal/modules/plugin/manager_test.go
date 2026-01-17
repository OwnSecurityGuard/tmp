package plugin

import (
	"fmt"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	m.Register("tya", "E:\\reply\\backend\\cmd\\test\\test.exe")
	err := m.Load("tya")
	d, err := m.Decode("tya", true, []byte("sds"))
	fmt.Println(err)
	fmt.Println(d)
	//for i := 0; i < 10; i++ {
	//	if i == 5 {
	//		m.Load("tya")
	//	}
	//	//a, _ := m.Decode("tya", []byte("sds"))
	//	//fmt.Println(a)
	//}

	//fmt.Println(a)
	//err = m.Load("tya2", "E:\\reply\\backend\\internal\\modules\\tmp\\tmp2.exe")
	//
	//a, _ = m.Decode("tya2", []byte("sds"))
	//fmt.Println(a)
}
