package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var letterRunes = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// set rand.seed
func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandStringRunes : Production random number id
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func sendSMS(ip, x, port string) (robots []byte) {
	url := "http://" + ip + "/sendsms?username=smsuser&password=smspwd&phonenumber=411&timeout=60&port=" + port + "&id=" + x + "&message=simbank-vpn-test-" + x
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	robots, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return

}

func getSmsReport(ip, x string) (robots []byte) {
	url := "http://" + ip + "/smsstatus?phonenumber=411&id=" + x
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	robots, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return
}

func getAllstatus(ip string) (robots []byte) {
	url := "http://admin:Admin123@" + ip + "/service?action=chan_state&username=admin&password=Admin123"
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	robots, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return

}

func checkImsi(ip, port string) (robots []byte) {
	url := "http://" + ip + "/simstatus?port=" + port
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	robots, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return
}

func call(wg *sync.WaitGroup, host, port string) {
	nowTime := time.Now().Format("2006-01-02 15:04:05")
	logFile := fmt.Sprintf("result-each/send-sms-%s.log", port)
	x := RandStringRunes(10)
	getAllState := getAllstatus(host)
	str := fmt.Sprintf("___%s___\n",nowTime)
	str += fmt.Sprintf("--> Get ALL Port Status:\n%s\n", getAllState)

	sendRes := sendSMS(host, x, port) // use http sms api send sms
	str += fmt.Sprintf("--> Send SMS:\n%s\n", sendRes)
	str += fmt.Sprintln("--> Get SMS report:")
getres:
	for i := 0; i < 5; i++ {
		time.Sleep(10 * time.Second) // wite for sms sending
		res := getSmsReport(host, x) // check id sms
		if len(res) > 40 {
			str += fmt.Sprintf("%s\n", res)
			break getres
		}
	}
	checkImsiRes := checkImsi(host, port)
	str += fmt.Sprintf("--> Check <%s> IMSI Result:\n%s\n", port, checkImsiRes)
	str += fmt.Sprintf("--> SMS ID:\n%s\n", x) // id number
	str += "\n<<*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*>>\n\n"
	info := []byte(str)

	_, err := os.Stat("result-each") // 判断文件夹是否存在,并创建
	if os.IsNotExist(err) {
		_ = os.Mkdir("result-each", os.ModePerm)
	}

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) // write or append info to file
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err := f.Write(info); err != nil {
		log.Fatal(err)
	}
	// // ioutil.WriteFile(logFile, info, 0644) // 快速写文件
	wg.Done()
}

func main() {
	var wg sync.WaitGroup
	host := ""
	port := ""
	flag.StringVar(&host, "host", "172.16.6.142", "GW-host...")
	flag.StringVar(&port, "port", "1", `GW-port...
	Multiple ports can be passed in separated by spaces 
	The script will operate in parallel
	The result is saved in the result-each folder`)
	flag.Parse()
	PL := strings.Split(port, " ")
	fmt.Println("Running, please wait...")
	for _, p := range PL {
		wg.Add(1)
		go call(&wg, host, p)
	}
	wg.Wait()
	fmt.Println("all", PL, "port sending is  Completed...")
}
