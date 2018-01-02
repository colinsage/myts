package tests

import (
	"testing"
	"net/http"
	"io/ioutil"
	"net/url"
	"strings"
	"fmt"
)

func TestCluster(t *testing.T){
	TestDropDatabaseNode1(t)

	TestCreateDatabaseNode1(t)

	TestWriteDataNode1(t)

	TestReadDataDiffrentNode(t)
}

func TestDropDatabaseNode1(t *testing.T){
	data := make(url.Values)
	data["q"] = []string{"drop DATABASE mydb"}
	resp , err := http.PostForm("http://localhost:7086/query"  ,data)

	if err != nil {
		t.Error("drop database failed. " + err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if "{\"results\":[{\"statement_id\":0}]}" == string(body) {
		t.Error(" drop database failed ." + string(body))
	}

}
func TestCreateDatabaseNode1(t *testing.T){
	//create database
	data := make(url.Values)
	data["q"] = []string{"CREATE DATABASE mydb"}
	resp , err := http.PostForm("http://localhost:7086/query"  ,data)
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

func TestWriteDataNode1(t *testing.T) {

	resp, err := http.Post("http://localhost:7086/write?db=mydb", "application/octet-stream ", strings.NewReader("cpu_load_long,host=server10,region=us-west value=12 1514055567000000000"))

	if err!= nil {
		t.Error("write failed. " + err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func TestReadDataDiffrentNode(t *testing.T){
	data := make(url.Values)
	data["db"] = []string{"mydb"}
	data["q"] = []string{"SELECT * FROM \"cpu_load_long\""}

	resp1, err1 := http.PostForm("http://localhost:7086/query", data)
	if err1!= nil {
		t.Error("read 7086 failed. " + err1.Error())
	}
	defer resp1.Body.Close()
	body1, _ := ioutil.ReadAll(resp1.Body)

	resp2, err2 := http.PostForm("http://localhost:8086/query", data)
	if err2!= nil {
		t.Error("read 8086 failed. " + err2.Error())
	}
	defer resp1.Body.Close()
	body2, _ := ioutil.ReadAll(resp2.Body)

	if string(body1) != string(body2) {
		t.Error("read result from 7086 and 8086 is diffent")
	}


}