package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
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
	logFile := fmt.Sprintf("result/send-sms-%s.log", port)
	x := RandStringRunes(10)
	getAllState := getAllstatus(host)
	str := fmt.Sprintf("--> Get ALL Port Status:\n%s\n", getAllState)

	sendRes := sendSMS(host, x, port) // use http sms api send sms
	str += fmt.Sprintf("--> Send SMS:\n%s\n", sendRes)

	str += fmt.Sprintln("--> Get SMS report:")
getres:
	for i := 0; i < 5; i++ {
		time.Sleep(10 * time.Second) // wite for sms sending
		res := getSmsReport(host, x) // check id sms
		if len(res) > 30 {
			str += fmt.Sprintf("%s\n", res)
			break getres
		}
	}
	checkImsiRes := checkImsi(host, port)
	str += fmt.Sprintf("--> Check <%s> IMSI Result:\n%s\n", port, checkImsiRes)
	str += fmt.Sprintf("--> SMS ID:\n%s\n", x) // id number
	str += "\n<<*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*_*>>\n\n"
	info := []byte(str)

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) // write or append info to file
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err := f.Write(info); err != nil {
		log.Fatal(err)
	}

	// ioutil.WriteFile(logFile, info, 0644) // 快速写文件
	wg.Done()
}

func main() {
	host := "172.16.6.142"
	var wg sync.WaitGroup
	count := 0
	for i := 1; i <= 16; i++ {
		if i == 7 || i == 13 {
			continue
			wg.Add(1)
			go call(&wg, host, strconv.Itoa(i))
			count++
		} else {
			wg.Add(1)
			go call(&wg, host, strconv.Itoa(i))
			count++
		}
	}
	wg.Wait()
	fmt.Println("all ", count, "port sending is  Completed...")
}
