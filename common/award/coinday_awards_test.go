/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package award

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

//获取币龄
func TestCoinDay(t *testing.T) {
	var (
		startTime1, _ = time.Parse("2006-01-02 15:04:05", "2018-12-01 00:00:00")
		//startTime2, _ = time.Parse("2006-01-02 15:04:05", "2018-11-26 01:00:00")
		//startTime3, _ = time.Parse("2006-01-02 15:04:05", "2018-11-25 02:00:00")
		//startTime4, _ = time.Parse("2006-01-02 15:04:05", "2018-11-24 03:00:00")
		////startTime5, _ = time.Parse("2006-01-02 15:04:05", "2007-01-02 00:00:00")
		endTime1, _ = time.Parse("2006-01-02 15:04:05", "2018-12-05 07:57:13")
		//endTime2, _ = time.Parse("2006-01-02 15:04:05", "2018-11-27 01:00:00")
		//endTime3, _ = time.Parse("2006-01-02 15:04:05", "2018-11-27 03:00:00")
		//endTime4, _ = time.Parse("2006-01-02 15:04:05", "2018-11-27 23:00:00")
	)
	tests := []struct {
		startTime time.Time
		endTime   time.Time
		want      int64
	}{
		{
			startTime: startTime1.UTC(),
			endTime:   endTime1,
			want:      40000,
		},
		//{
		//	startTime: startTime2.UTC().Unix(),
		//	endTime:   endTime2,
		//	want:      1,
		//},
		//{
		//	startTime: startTime3.UTC().Unix(),
		//	endTime:   endTime3,
		//	want:      2,
		//},
		//{
		//	startTime: startTime4.UTC().Unix(),
		//	endTime:   endTime4,
		//	want:      3,
		//},
		//{
		//	startTime: startTime5.UTC().Unix(),
		//	endTime:   time.Now().UTC(),
		//	want:      4349,
		//},
	}
	for i, test := range tests {
		duration := GetCoinDay(10000, test.startTime, test.endTime)
		if int64(duration) != test.want {
			t.Errorf("the %d failed,want %d but get %d", i, test.want, duration)
		} else {
			t.Logf("the %d succeeded,want %d and get %d", i, test.want, duration)
		}
	}
}

func TestCalculateAwardsForDepositContractNodes(t *testing.T) {
	startTime, _ := time.Parse("2006-01-02 15:04:05", "2018-12-01 00:00:00")
	endTime, _ := time.Parse("2006-01-02 15:04:05", "2018-12-05 07:57:13")
	//距离现在天数：4349
	//获取币龄 余额：1000
	fmt.Println(time.Now().UTC().Unix())
	fmt.Println((endTime.Unix() - startTime.Unix()) / 60 / 60 / 24)
	fmt.Println((endTime.Unix()/60 - startTime.Unix()/60) / 60 / 24)
	fmt.Println(startTime.UTC().Unix())
	fmt.Println(endTime.UTC().Unix() / 60 / 60 / 24)

	coinDayUint64 := GetCoinDay(10000, startTime.UTC(), endTime)
	//币龄：4349000
	//获取币龄收益
	awards := CalculateAwardsForDepositContractNodes(coinDayUint64,"0.02")
	if awards != 2 {
		t.Errorf("failed,want 2,but get %d", awards)
	} else {
		t.Logf("succeeded,want 2 and get %d", awards)
	}
}

func TestTime(t *testing.T) {
	//txid := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	//fmt.Println(txid)
	//fmt.Println(time.Now().UTC().Unix())
	//time.Sleep(1 * time.Second)
	//txid = strconv.FormatInt(time.Now().UTC().Unix(), 10)
	//fmt.Println(txid)

	tx := fmt.Sprintf("%08v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(100000000))
	fmt.Println(tx)
	time.Sleep(1 * time.Second)
	tx = fmt.Sprintf("%08v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(100000000))
	fmt.Println(tx)
	time.Sleep(1 * time.Second)
	tx = fmt.Sprintf("%08v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(100000000))
	fmt.Println(tx)
}
