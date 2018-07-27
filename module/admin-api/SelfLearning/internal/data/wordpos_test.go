package data

import (
	"fmt"
	"os"
	"testing"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"
)

func TestWordpos(t *testing.T) {
	util.LogInit("TEST", os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	nluURL := "http://172.16.101.47:13901"
	testQuestions := []string{
		"APP交易指南—如何买入/卖出",
		"APP内有止盈止损的功能吗",
		"APP新版本有什么不一样？",
		"APP交易指南—如何持仓查询",
		"APP里从那里推荐好友注册帐户",
		"APP在线业务办理—在线开通劵E融",
		"APP上如何查大盘的平均市盈率？",
		"APP在线业务办理—如何修改证件有效期",
		"APP预约开通融资融券，申请时提示未开通交易，这是啥情况？",
		"App6.3什么时候出",
		"App首页没有重置密码这一项",
		"App自助开户可以的不？需要填写推荐经纪人的不？",
		"A股万分之二点2五的交易佣金是否包括规费？什么是规费？",
		"A股手续费有那些",
		"A股股东卡尚未指定交易",
		"B转H股，如果成功在香港上市，境内原持有B股的股民该如何操作",
		"B股T+1回转交易与T+3交收是指什么",
		"B股账户",
		"B股的概念",
		"B股交易时间",
		"B股交易规则",
		"B股账号开通",
		"B股转托管何时到账",
		"B股交易所收费明细表有哪些",
		"B股市值不计入打新额度计算吧",
		"D指标怎样选股",
		"D信号怎么没有显示",
		"D信号我要退掉后面11个月",
		"D信号用户如何接受到？",
		"D信号，指标怎么安装，使用方法，",
		"ETF交易在哪个菜单栏？",
		"E融资怎么还钱",
		"F6打开自选股后，如何全屏显示所有自选股",
		"FuQu",
		"H股是指什么",
		"JCDJ111",
		"Klnoqu",
		"K线图不显示",
		"K线图的画法",
		"Lev2电脑可用吗？手机购买。",
		"Maybe we never really know how good we have it, until it’s gone",
		"OBV线使用方法",
		"OVL是什么意思",
		"PA18网上开户支持的三方存管银行",
		"PAD炒股软件下载",
		"PC端软件怎么下不了",
		"PC端怎么操作风险评测",
		"PC上的金融终端下载安装后无法运行",
		"ST股怎样开通",
		"Sometimes, the same thing, we can go to the comfort of others, but failed to convince yourself.",
		"T+1是什么",
		"T+1到账是什么意思",
		"T日赎回，资金什么时候到账？",
		"V1P是收费的吧",
		"WIFI的环境下，理财页面打开显示空白",
		"a0d9673a03af040ae7bd82fcfc79d378",
		"a11de3bbf9ed77ec58c19ab9ef08c5c1",
		"a12d8f775be0bbdb99209b6975e7935d",
		"a89432f423f79f4b7403ac7e28efafea",
		"ac3adaf5bc41af3d212a9dd5f11951a3",
		"af65f9f3eb1f63c184738c64f5593001",
	}
	testQuestionIDs := make([]uint64, len(testQuestions))
	for idx := range testQuestionIDs {
		testQuestionIDs[idx] = uint64(idx) + 1
	}
	start := time.Now()
	var nativeLog NativeLog
	nativeLog.Init()
	nativeLog.GetWordPos(nluURL, testQuestions, testQuestionIDs)
	util.LogInfo.Printf("Test ends. Calculate [%v] questions in [%v]s\n", len(testQuestions), time.Since(start).Seconds())
	for i := 0; i < len(nativeLog.Logs); i++ {
		datItem := nativeLog.Logs[i]
		fmt.Println(datItem.Tokens)
		fmt.Println(datItem.KeyWords)
		fmt.Println()
	}
}
