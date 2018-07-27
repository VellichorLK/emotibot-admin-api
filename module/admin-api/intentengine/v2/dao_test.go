package v2

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"testing"

	"emotibot.com/emotigo/module/admin-api/util"
	"github.com/lestrrat-go/test-mysqld"
)

var testDao intentDaoV2
var mysqld *mysqltest.TestMysqld

func setup() error {
	var err error
	if testDao.db == nil {
		mysqld, err = mysqltest.NewMysqld(nil)
		if err != nil {
			log.Fatalln("Fail to start test mysqld:", err.Error())
		}

		db, err := sql.Open("mysql", mysqld.Datasource("emotibot", "", "", 0))
		if err != nil {
			log.Fatalln("Fail to open mysql:", err.Error())
		}
		err = setupTestDB(db)
		if err != nil {
			log.Fatalln("Fail to setup mysql:", err.Error())
		}
		testDao.db = db
	}
	util.LogInit("TEST")
	return nil
}

func setupTestDB(db *sql.DB) error {
	db.Exec("CREATE DATABASE emotibot CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	db.Exec("USE emotibot")
	_, err := db.Exec("CREATE TABLE `intent_train_sets` ( `id` int(11) NOT NULL AUTO_INCREMENT, `sentence` varchar(256) COLLATE utf8_unicode_ci NOT NULL, `intent` int(11) NOT NULL, `type` int(1) NOT NULL DEFAULT '0', PRIMARY KEY (`id`), KEY `intent_id` (`intent`)) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;")
	if err != nil {
		fmt.Println("Create intent_train_sets meet error: ", err.Error())
		return err
	}
	_, err = db.Exec("CREATE TABLE `intent_versions` ( `intent_version_id` int(11) NOT NULL AUTO_INCREMENT, `app_id` varchar(128) COLLATE utf8_unicode_ci NOT NULL, `ie_model_id` varchar(128) COLLATE utf8_unicode_ci DEFAULT NULL, `re_model_id` varchar(128) COLLATE utf8_unicode_ci DEFAULT NULL, `orig_file_name` varchar(256) COLLATE utf8_unicode_ci NOT NULL, `file_name` varchar(256) COLLATE utf8_unicode_ci NOT NULL, `in_used` tinyint(1) NOT NULL DEFAULT '0', `created_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP, `updated_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (`intent_version_id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;")
	if err != nil {
		fmt.Println("Create intent_versions meet error: ", err.Error())
		return err
	}
	_, err = db.Exec("CREATE TABLE `intents` ( `id` int(11) NOT NULL AUTO_INCREMENT, `appid` varchar(128) COLLATE utf8_unicode_ci NOT NULL DEFAULT '', `name` varchar(256) COLLATE utf8_unicode_ci NOT NULL, `version` int(11) DEFAULT NULL, `updatetime` int(20) NOT NULL, PRIMARY KEY (`id`), KEY `intent_version_id` (`version`)) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;")
	if err != nil {
		fmt.Println("Create intent_versions meet error: ", err.Error())
		return err
	}
	_, err = db.Exec("INSERT INTO `intents` (`id`, `appid`, `name`, `version`, `updatetime`) VALUES (1, 'test','记支出',NULL,0), (2, 'test','记收入',NULL,0), (3, 'test','查支出',NULL,0), (4, 'test','查汇总',NULL,0), (5, 'test','查收入',NULL,0), (6, 'test','other',NULL,0), (7, 'test','other',1,0);")
	if err != nil {
		fmt.Println("Insert intents meet error: ", err.Error())
		return err
	}
	_, err = db.Exec("INSERT INTO `intent_train_sets` (`id`, `sentence`, `intent`, `type`) VALUES (1, '支出1', 1, 0), (2, '不出I', 1, 1)")
	if err != nil {
		fmt.Println("Insert intent_train_sets meet error: ", err.Error())
		return err
	}
	return nil
}

func teardown() {
	if testDao.db != nil {
		testDao.db.Close()
		testDao.db = nil
	}
	if mysqld != nil {
		mysqld.Stop()
		mysqld = nil
	}
}

func TestGetIntents(t *testing.T) {
	setup()
	intents, err := testDao.GetIntents("test", nil, "")
	if err != nil {
		t.Error(
			"For", "appid [test], version [nil]",
			"expected", nil,
			"got", err,
		)
	}
	if len(intents) != 6 {
		t.Error(
			"For", "Count",
			"expected", 6,
			"got", len(intents),
		)
	}

	intents, err = testDao.GetIntents("test", nil, "支出")
	if err != nil {
		logError(t, "Get intents with keyword", nil, err)
	}
	if len(intents) != 2 {
		logError(t, "Get intents with keyword 支出", 2, len(intents))
	}
	var intent *IntentV2
	for idx := range intents {
		if intents[idx].Name == "记支出" {
			intent = intents[idx]
		}
	}
	if intent == nil {
		logError(t, "Get intents with keyword 支出, check search result", "valid pointer", nil)
	} else {
		if intent.PositiveCount != 1 {
			logError(t, "Get intents with keyword 支出, check positive sentence count", 1, intent.PositiveCount)
		}
		if intent.NegativeCount != 0 {
			logError(t, "Get intents with keyword 支出, check negative sentence count", 0, intent.NegativeCount)
		}
	}

	intents, err = testDao.GetIntents("test", nil, "不出i")
	if err != nil {
		logError(t, "Get intents with keyword", nil, err)
	}
	if len(intents) != 1 {
		logError(t, "Get intents with keyword 不出i", 1, len(intents))
	}
	for idx := range intents {
		if intents[idx].Name == "记支出" {
			intent = intents[idx]
		}
	}
	if intent == nil {
		logError(t, "Get intents with keyword 不出i, check search result", "valid pointer", nil)
	} else {
		if intent.PositiveCount != 0 {
			logError(t, "Get intents with keyword 不出i, check positive sentence count", 0, intent.PositiveCount)
		}
		if intent.NegativeCount != 1 {
			logError(t, "Get intents with keyword, 不出i check negative sentence count", 1, intent.NegativeCount)
		}
	}

	intents, err = testDao.GetIntents("csbot", nil, "")
	if err != nil {
		t.Error(
			"For", "appid [csbot], version [nil]",
			"expected", nil,
			"got", err,
		)
	}
	if len(intents) != 0 {
		t.Error(
			"For", "Count",
			"expected", 0,
			"got", len(intents),
		)
	}

	version := 2
	intents, err = testDao.GetIntents("test", &version, "")
	if err != sql.ErrNoRows {
		t.Error(
			"For", "appid [test], version [1]",
			"expected", sql.ErrNoRows,
			"got", nil,
		)
	}
	teardown()
}

func TestGetIntent(t *testing.T) {
	setup()
	intent, err := testDao.GetIntent("test", 1, "")
	if err != nil {
		logError(t, "Get intent of appid [test], version [nil]", nil, err)
	}
	if intent.PositiveCount != 1 {
		logError(t, "Get positive count of appid [test], version [nil]", 1, intent.PositiveCount)
	}
	if (*intent.Positive)[0].Content != "支出1" {
		logError(t, "Get positive content of appid [test], version [nil]", "支出1", (*intent.Positive)[0].Content)
	}
	if intent.NegativeCount != 1 {
		logError(t, "Get negative count of appid [test], version [nil]", 1, intent.NegativeCount)
	}
	if (*intent.Negative)[0].Content != "不出I" {
		logError(t, "Get positive content of appid [test], version [nil]", "不出I", (*intent.Negative)[0].Content)
	}

	intent, err = testDao.GetIntent("test", 1, "支出")
	if err != nil {
		logError(t, "Get intent of appid [test], version [nil], keyword [支出]", nil, err)
	}
	if intent.PositiveCount != 1 {
		logError(t, "Get positive count of appid [test], version [nil], keyword [支出]", 1, intent.PositiveCount)
	}
	if (*intent.Positive)[0].Content != "支出1" {
		logError(t, "Get positive content of appid [test], version [nil], keyword [支出]", "支出1", (*intent.Positive)[0].Content)
	}
	if intent.NegativeCount != 0 {
		logError(t, "Get negative count of appid [test], version [nil], keyword [支出]", 0, intent.NegativeCount)
	}
	teardown()
}

func TestAddIntent(t *testing.T) {
	setup()
	intent, err := testDao.AddIntent("test", "addIntent",
		[]string{"testPositiveAdd"},
		[]string{"testNegativeAdd"})
	if err != nil {
		t.Error(
			"Add intent fail",
			"expected", nil,
			"got", err,
		)
	}
	if intent.Name != "addIntent" {
		t.Error(
			"Added intent name diff",
			"expected", "addIntent",
			"got", intent.Name,
		)
	}
	if intent.Name != "addIntent" {
		t.Error(
			"Added intent name diff",
			"expected", "addIntent",
			"got", intent.Name,
		)
	}
	if intent.PositiveCount != 1 || intent.Positive == nil || len(*intent.Positive) != 1 ||
		(*intent.Positive)[0].Content != "testPositiveAdd" {
		if intent.Positive == nil || len(*intent.Positive) != 1 {
			t.Error("Added intent positive sentence fail")
		} else {
			t.Error(
				"Added intent positive sentence diff",
				"expected", "addIntent",
				"got", (*intent.Positive)[0].Content,
			)
		}
	}
	if intent.NegativeCount != 1 || intent.Negative == nil || len(*intent.Negative) != 1 ||
		(*intent.Negative)[0].Content != "testNegativeAdd" {
		if intent.Negative == nil || len(*intent.Negative) != 1 {
			t.Error("Added intent negative sentence fail")
		} else {
			t.Error(
				"Added intent negative sentence diff",
				"expected", "addIntent",
				"got", (*intent.Negative)[0].Content,
			)
		}
	}

	resultIntent, err := testDao.GetIntent("test", intent.ID, "")
	if err != nil {
		t.Error(
			"Get added intent fail",
			"expected", nil,
			"got", err,
		)
	}

	intentBytes, _ := json.Marshal(intent)
	resultBytes, _ := json.Marshal(resultIntent)
	if !bytes.Equal(intentBytes, resultBytes) {
		t.Error(
			"New intent different with return",
			"expected", intent.Name,
			"got", resultIntent.Name,
		)
	}

	teardown()
}

func TestUpdateIntent(t *testing.T) {
	var err error
	setup()
	t.Run("Error test", func(t *testing.T) {
		err = testDao.ModifyIntent("csbot", 1, "123", nil, nil)
		if err != sql.ErrNoRows {
			logError(t, "Modify invalid intent", sql.ErrNoRows, err)
		}
		err = testDao.ModifyIntent("test", 7, "newName", nil, nil)
		if err != ErrReadOnlyIntent {
			logError(t, "Modify commited intent", ErrReadOnlyIntent, err)
		}
	})

	t.Run("Update test", func(t *testing.T) {
		err = testDao.ModifyIntent("test", 1, "记支出new", nil, nil)
		if err != nil {
			logError(t, "Modify intent name only", nil, err)
		}

		err = testDao.ModifyIntent("test", 1, "记支出new", []*SentenceV2WithType{
			&SentenceV2WithType{SentenceV2{0, "支出2"}, 0},
			&SentenceV2WithType{SentenceV2{0, "支出3"}, 0},
			&SentenceV2WithType{SentenceV2{1, "支出11"}, 0},
		}, nil)
		if err != nil {
			logError(t, "Modify intent, update only", nil, err)
		}

		err = testDao.ModifyIntent("test", 1, "记支出new", nil, []int64{2})
		if err != nil {
			logError(t, "Modify intent, delete only", nil, err)
		}

		intent, err := testDao.GetIntent("test", 1, "")
		if err != nil {
			logError(t, "Get modified intent", nil, err)
		}
		if intent.Name != "记支出new" {
			logError(t, "Name modify fail", nil, err)
		}
		if intent.PositiveCount != 3 {
			logError(t, "Positive add fail", 3, intent.PositiveCount)
		}
		if intent.Positive == nil {
			logError(t, "Get Positive fail after update", "valid pointer", nil)
		}
		except := []string{"支出11", "支出2", "支出3"}
		get := []string{}
		for _, sentence := range *intent.Positive {
			get = append(get, sentence.Content)
		}

		sort.Sort(sort.StringSlice(except))
		sort.Sort(sort.StringSlice(get))
		if strings.Join(except, ",") != strings.Join(get, ",") {
			logError(t, "Positive modify fail", strings.Join(except, ","), strings.Join(get, ","))
		}

		if intent.NegativeCount != 0 {
			logError(t, "Delete intent sentence fail", 0, intent.NegativeCount)
		}
		if intent.Negative == nil {
			logError(t, "Get Negative fail after update", "valid pointer", nil)
		}
	})

	teardown()
}

func TestDeleteIntent(t *testing.T) {
	setup()

	err := testDao.DeleteIntent("test", 7)
	if err != ErrReadOnlyIntent {
		logError(t, "Delete read only", ErrReadOnlyIntent, err)
	}

	err = testDao.DeleteIntent("test", 10)
	if err != nil {
		logError(t, "Delete not exised", nil, err)
	}

	err = testDao.DeleteIntent("test", 6)
	if err != nil {
		logError(t, "Test delete", nil, err)
	}
	_, err = testDao.GetIntent("test", 6, "")
	if err != sql.ErrNoRows {
		logError(t, "Test get after delete", sql.ErrNoRows, err)
	}

	teardown()
}

func logError(t *testing.T, item string, except interface{}, get interface{}) {
	t.Errorf("TESTING [%s], except [%+v], get[%+v]\n", item, except, get)
}
