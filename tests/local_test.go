package tests

import (
	"testing"
	"net/http"
	"io/ioutil"
	"net/url"
	"strings"
	"fmt"
)

func TestLocal(t *testing.T){
	TestDropDatabaseNode(t)

	TestCreateDatabaseNode(t)

	TestWriteDataNode(t)

	TestReadData(t)
}

func TestDropDatabaseNode(t *testing.T){
	data := make(url.Values)
	data["q"] = []string{"drop DATABASE mydb"}
	resp , err := http.PostForm("http://localhost:6086/query"  ,data)

	if err != nil {
		t.Error("drop database failed. " + err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if "{\"results\":[{\"statement_id\":0}]}" == string(body) {
		t.Error(" drop database failed ." + string(body))
	}

}
func TestCreateDatabaseNode(t *testing.T){
	//create database
	data := make(url.Values)
	data["q"] = []string{"CREATE DATABASE mydb"}
	resp , err := http.PostForm("http://localhost:6086/query"  ,data)
	//

	if err != nil {
		t.Error("create database failed. " + err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if "{\"results\":[{\"statement_id\":0}]}" == string(body) {
		t.Error(" create database failed ." + string(body))
	}
}

func TestWriteDataNode(t *testing.T) {

	resp, err := http.Post("http://localhost:6086/write?db=mydb", "application/octet-stream ", strings.NewReader("cpu_load_long,host=server10,region=us-west value=12 1514055567000000000"))

	if err!= nil {
		t.Error("write failed. " + err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func TestReadData(t *testing.T){
	data := make(url.Values)
	data["db"] = []string{"mydb"}
	data["q"] = []string{"SELECT * FROM \"cpu_load_long\""}

	resp1, err1 := http.PostForm("http://localhost:6086/query", data)
	if err1!= nil {
		t.Error("read 7086 failed. " + err1.Error())
	}
	defer resp1.Body.Close()
	body1, _ := ioutil.ReadAll(resp1.Body)
	fmt.Println(string(body1))
}