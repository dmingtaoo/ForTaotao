package router

import (
	"dpchain/api"
	"dpchain/core/consensus"
	"dpchain/core/eles"
	"dpchain/dper/transactionCheck"
	"dpchain/ginHttp/pkg/app"
	e "dpchain/ginHttp/pkg/error"
	loglogrus "dpchain/log_logrus"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"sync"

	"github.com/astaxie/beego/validation"
	"github.com/gin-gonic/gin"
)

func BackAccountList(ds *api.DperService) func(*gin.Context) {
	return func(c *gin.Context) {
		loglogrus.Log.Infof("Http: Get user request -- BackAccountList\n")
		appG := app.Gin{c}
		data, err := ds.BackListAccounts()
		if err != nil {
			loglogrus.Log.Warnf("Http: Reply to user request -- BackAccountList  warn:%s\n", err)
			appG.Response(http.StatusOK, e.ERROR, err)
			return
		} else {
			appG.Response(http.StatusOK, e.SUCCESS, data)
		}
		loglogrus.Log.Infof("Http: Reply to user request -- BackAccountList  succeed!\n")
	}
}

func CreateNewAccount(ds *api.DperService) func(*gin.Context) {
	return func(c *gin.Context) {
		loglogrus.Log.Infof("Http: Get user request -- CreateNewAccount\n")
		appG := app.Gin{c}
		password := c.PostForm("password")

		commandStr := "createNewAccount"

		if password != "" {
			commandStr += (" -p " + password)
		}

		account, err := ds.CreateNewAccount(commandStr)
		if err != nil {
			loglogrus.Log.Warnf("Http: Reply to user request -- CreateNewAccount  warn:%s\n", err)
			appG.Response(http.StatusOK, e.ERROR, err)
			return
		} else {
			appG.Response(http.StatusOK, e.SUCCESS, fmt.Sprintf("%x", account))
		}
		loglogrus.Log.Infof("Http: Reply to user request -- CreateNewAccount  succeed!\n")
	}

}

// 十六进制字符串格式的account
func UseAccount(ds *api.DperService) func(*gin.Context) {
	return func(c *gin.Context) {
		appG := app.Gin{c}
		account := c.PostForm("account")
		password := c.PostForm("password")
		valid := validation.Validation{}
		valid.Required(account, "account").Message("账户Hash不能为空")
		if valid.HasErrors() {
			app.MarkErrors("", valid.Errors)
			appG.Response(http.StatusOK, e.INVALID_PARAMS, nil)
			return
		}
		loglogrus.Log.Infof("Http: Get user request -- UseAccount  account:%s\n", account)

		commandStr := "useAccount " + account

		if password != "" {
			commandStr += (" -p " + password)
		}

		err := ds.UseAccount(commandStr)
		if err != nil {
			loglogrus.Log.Warnf("Http: Reply to user request -- UseAccount  warn:%s\n", err)
			appG.Response(http.StatusOK, e.ERROR, err)
			return
		} else {
			appG.Response(http.StatusOK, e.SUCCESS, nil)
		}
		loglogrus.Log.Infof("Http: Reply to user request -- UseAccount  succeed!\n")
	}

}

func CurrentAccount(ds *api.DperService) func(*gin.Context) {
	return func(c *gin.Context) {
		loglogrus.Log.Infof("Http: Get user request -- CurrentAccount\n")
		appG := app.Gin{c}
		infoStr := ds.CurrentAccount()

		appG.Response(http.StatusOK, e.SUCCESS, infoStr)
		loglogrus.Log.Infof("Http: Reply to user request -- CurrentAccount  succeed!\n")
	}

}

// on:开启  off:关闭
func OpenTxCheckMode(ds *api.DperService) func(*gin.Context) {
	return func(c *gin.Context) {
		appG := app.Gin{c}
		mode := c.PostForm("mode")
		valid := validation.Validation{}
		valid.Required(mode, "commandStr").Message("on:开启交易上链检查 off:关闭交易上链检查")
		if valid.HasErrors() {
			app.MarkErrors("OpenTxCheckMode", valid.Errors)
			appG.Response(http.StatusOK, e.INVALID_PARAMS, nil)
			return
		}
		loglogrus.Log.Infof("Http: Get user request -- OpenTxCheckMode  mode:%s\n", mode)
		commandStr := "txCheckMode " + mode

		err := ds.OpenTxCheckMode(commandStr)
		if err != nil {
			loglogrus.Log.Infof("Http: Reply to user request -- OpenTxCheckMode  warn:%s\n", err)
			appG.Response(http.StatusOK, e.ERROR, err)
			return
		} else {
			appG.Response(http.StatusOK, e.SUCCESS, nil)
		}
		loglogrus.Log.Infof("Http: Reply to user request -- OpenTxCheckMode  succeed!\n")
	}

}

// contractAddr:合约地址
// functionAddr:参数地址
// args: 参数列表,以空格分割
func SolidInvoke(ds *api.DperService) func(*gin.Context) {
	return func(c *gin.Context) {
		appG := app.Gin{c}
		contractAddr := c.PostForm("contractAddr")
		functionAddr := c.PostForm("functionAddr")
		args := c.PostForm("args")
		valid := validation.Validation{}
		valid.Required(contractAddr, "contractAddr").Message("合约地址不能为空")
		valid.Required(functionAddr, "functionAddr").Message("函数地址不能为空")
		valid.Required(args, "args").Message("参数不能为空")
		if valid.HasErrors() {
			app.MarkErrors("SolidInvoke", valid.Errors)
			appG.Response(http.StatusOK, e.INVALID_PARAMS, nil)
			return
		}
		loglogrus.Log.Infof("Http: Get user request -- SolidInvoke  contractAddr:%s , functionAddr:%s , args:%s\n", contractAddr, functionAddr, args)
		commandStr := "solidInvoke " + contractAddr + " " + functionAddr + " -args " + args

		receipt, err := ds.SolidInvoke(commandStr)
		if err != nil {
			loglogrus.Log.Warnf("Http: Reply to user request -- SolidInvoke  warn:%s\n", err)
			appG.Response(http.StatusOK, e.ERROR, err)
			return
		} else {

			if reflect.DeepEqual(receipt, transactionCheck.CheckResult{}) {
				appG.Response(http.StatusOK, e.SUCCESS, nil)
			} else {
				type ResponseReceipt struct {
					TxID   string `json:"Transaction ID"`
					Valid  bool   `json:"Valid"`
					Result string `json:"Transaction Results"`
					Delay  string `json:"Consensus Delay"`
				}
				var r ResponseReceipt = ResponseReceipt{
					TxID:   fmt.Sprintf("%x", receipt.TransactionID),
					Valid:  receipt.Valid,
					Result: fmt.Sprintf("%s", receipt.Result),
					Delay:  fmt.Sprintf("%d ms", receipt.Interval),
				}
				appG.Response(http.StatusOK, e.SUCCESS, r)
			}

		}
		loglogrus.Log.Infof("Http: Reply to user request -- SolidInvoke  succeed!\n")
	}

}

// contractAddr:合约地址
// functionAddr:参数地址
// args: 参数列表,以空格分割
func SolidCall(ds *api.DperService) func(*gin.Context) {
	return func(c *gin.Context) {
		appG := app.Gin{c}
		contractAddr := c.PostForm("contractAddr")
		functionAddr := c.PostForm("functionAddr")
		args := c.PostForm("args")
		valid := validation.Validation{}
		valid.Required(contractAddr, "contractAddr").Message("合约地址不能为空")
		valid.Required(functionAddr, "functionAddr").Message("函数地址不能为空")
		valid.Required(args, "args").Message("参数不能为空")

		if valid.HasErrors() {
			app.MarkErrors("SolidCall", valid.Errors)
			appG.Response(http.StatusOK, e.INVALID_PARAMS, nil)
			return
		}
		loglogrus.Log.Infof("Http: Get user request -- SolidCall  contractAddr:%s , functionAddr:%s , args:%s\n", contractAddr, functionAddr, args)
		commandStr := "solidCall " + contractAddr + " " + functionAddr + " -args " + args

		data, err := ds.SolidCall(commandStr)
		if err != nil {
			loglogrus.Log.Warnf("Http: Reply to user request -- SolidCall  warn:%s\n", err)
			appG.Response(http.StatusOK, e.ERROR, err)
			return
		} else {
			appG.Response(http.StatusOK, e.SUCCESS, data)
		}
		loglogrus.Log.Infof("Http: Reply to user request -- SolidCall  succeed!\n")
	}

}

// contractName:合约名
// functionName:参数名
// args: 参数列表,以空格分割
func SoftInvoke(ds *api.DperService) func(*gin.Context) {
	return func(c *gin.Context) {
		appG := app.Gin{c}
		contractName := c.PostForm("contractName")
		functionName := c.PostForm("functionName")
		args := c.PostForm("args")
		valid := validation.Validation{}
		valid.Required(contractName, "contractName").Message("合约名不能为空")
		valid.Required(functionName, "functionName").Message("函数名不能为空")
		valid.Required(args, "args").Message("参数不能为空")

		if valid.HasErrors() {
			app.MarkErrors("SoftInvoke", valid.Errors)
			appG.Response(http.StatusOK, e.INVALID_PARAMS, nil)
			return
		}
		loglogrus.Log.Infof("Http: Get user request -- SoftInvoke  contractName:%s , functionName:%s , args:%s\n", contractName, functionName, args)
		fmt.Printf("Http: Get user request -- SoftInvoke  contractName:%s , functionName:%s , args:%s\n", contractName, functionName, args)

		commandStr := "invoke " + contractName + " " + functionName + " -args " + args

		receipt, err := ds.SoftInvoke(commandStr)
		if err != nil {
			loglogrus.Log.Warnf("Http: Reply to user request -- SoftInvoke  warn:%s\n", err)
			appG.Response(http.StatusOK, e.ERROR, err)
			return
		} else {
			if reflect.DeepEqual(receipt, transactionCheck.CheckResult{}) {
				appG.Response(http.StatusOK, e.SUCCESS, nil)
			} else {
				type ResponseReceipt struct {
					TxID   string `json:"Transaction ID"`
					Valid  bool   `json:"Valid"`
					Result string `json:"Transaction Results"`
					Delay  string `json:"Consensus Delay"`
				}
				var r ResponseReceipt = ResponseReceipt{
					TxID:   fmt.Sprintf("%x", receipt.TransactionID),
					Valid:  receipt.Valid,
					Result: fmt.Sprintf("%s", receipt.Result),
					Delay:  fmt.Sprintf("%d ms", receipt.Interval.Milliseconds()),
				}
				appG.Response(http.StatusOK, e.SUCCESS, r)
			}
		}
		loglogrus.Log.Infof("Http: Reply to user request -- SoftInvoke  succeed!\n")
		fmt.Printf("Http: Reply to user request -- SoftInvoke  succeed!\n")
	}

}

func SoftInvokeQuery(ds *api.DperService) func(*gin.Context) {
	return func(c *gin.Context) {
		appG := app.Gin{c}
		contractName := c.PostForm("contractName")
		functionName := c.PostForm("functionName")
		args := c.PostForm("args")
		valid := validation.Validation{}
		valid.Required(contractName, "contractName").Message("合约名不能为空")
		valid.Required(functionName, "functionName").Message("函数名不能为空")
		valid.Required(args, "args").Message("参数不能为空")

		if valid.HasErrors() {
			app.MarkErrors("SoftInvoke", valid.Errors)
			appG.Response(http.StatusOK, e.INVALID_PARAMS, nil)
			return
		}
		loglogrus.Log.Infof("Http: Get user request -- SoftInvokeQuery contractName:%s , functionName:%s , args:%s\n", contractName, functionName, args)
		fmt.Printf("Http: Get user request -- SoftInvokeQuery contractName:%s , functionName:%s , args:%s\n", contractName, functionName, args)

		commandStr := "invoke " + contractName + " " + functionName + " -args " + args

		var wg sync.WaitGroup
		ds.SoftInvokeQuery(commandStr, wg)
		fmt.Printf("Http: Reply to user request -- SoftInvokeQuery  succeed!\n")
	}
}

func PublishTx(ds *api.DperService) func(*gin.Context) {
	return func(c *gin.Context) {
		// 1.解析 []tx 编码得到的byte流
		appG := app.Gin{c}
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			loglogrus.Log.Warnf("Http: PublishTx函数无法成功读取http请求,err:%v\n", err)
			appG.Response(http.StatusInternalServerError, e.ERROR, err)
			return
		} else {
			loglogrus.Log.Warnf("Http: PublishTx函数成功读取http请求,body比特流长度:%v\n", len(body))
		}
		// Parse函数
		txList, err := consensus.DeserializeTransactionsSeries(body)
		if err != nil {
			loglogrus.Log.Warnf("Http: PublishTx函数无法解析http请求内容,err:%v\n", err)
			appG.Response(http.StatusInternalServerError, e.ERROR, err)
			return
		} else {
			loglogrus.Log.Warnf("Http: PublishTx函数解析获得的交易数量:%v\n", len(txList))
		}

		// 2. 直接上传给consensusPromoter
		txListPtr := make([]*eles.Transaction, 0)
		for _, tx := range txList {
			txPtr := new(eles.Transaction)
			copyTransaction(tx, txPtr)
			txListPtr = append(txListPtr, txPtr)
		}

		err = ds.PublishTx(txListPtr)
		if err != nil {
			loglogrus.Log.Warnf("Http: PublishTx函数无法上传交易,err:%v\n", err)
			appG.Response(http.StatusOK, e.ERROR, err)
			return
		}
	}
}

func copyTransaction(src eles.Transaction, dst *eles.Transaction) {
	dst.TxID = src.TxID
	dst.Sender = src.Sender
	dst.Nonce = src.Nonce
	dst.Version = src.Version
	dst.LifeTime = src.LifeTime
	dst.Signature = append(dst.Signature, src.Signature...)
	dst.Contract = src.Contract
	dst.Function = src.Function
	dst.Args = append(dst.Args, src.Args...)
	dst.CheckList = append(dst.CheckList, src.CheckList...)
}

// contractName:合约名
// functionName:参数名
// args: 参数列表,以空格分割
func SoftCall(ds *api.DperService) func(*gin.Context) {
	return func(c *gin.Context) {
		appG := app.Gin{c}
		contractName := c.PostForm("contractName")
		functionName := c.PostForm("functionName")
		args := c.PostForm("args")
		valid := validation.Validation{}
		valid.Required(contractName, "contractName").Message("合约名不能为空")
		valid.Required(functionName, "functionName").Message("函数名不能为空")
		valid.Required(args, "args").Message("参数不能为空")

		if valid.HasErrors() {
			app.MarkErrors("SoftCall", valid.Errors)
			appG.Response(http.StatusOK, e.INVALID_PARAMS, nil)
			return
		}
		loglogrus.Log.Infof("Http: Get user request -- SoftCall  contractName:%s , functionName:%s , args:%s\n", contractName, functionName, args)

		commandStr := "call " + contractName + " " + functionName + " -args " + args

		data, err := ds.SoftCall(commandStr)
		if err != nil {
			loglogrus.Log.Warnf("Http: Reply to user request -- SoftCall  warn:%s\n", err)
			appG.Response(http.StatusOK, e.ERROR, err)
			return
		} else {
			appG.Response(http.StatusOK, e.SUCCESS, data)
		}

		loglogrus.Log.Infof("Http: Reply to user request -- SoftCall  succeed!\n")
	}

}

func BecomeBooter(ds *api.DperService) func(*gin.Context) {
	return func(c *gin.Context) {
		loglogrus.Log.Infof("Http: Get user request -- BecomeBooter")
		appG := app.Gin{c}
		err := ds.BecomeBooter()
		if err != nil {
			loglogrus.Log.Warnf("Http: Reply to user request -- BecomeBooter  warn:%s\n", err)
			appG.Response(http.StatusOK, e.ERROR, err)
			return
		} else {
			appG.Response(http.StatusOK, e.SUCCESS, nil)
		}
		loglogrus.Log.Infof("Http: Reply to user request -- BecomeBooter  succeed!\n")
	}

}

func BackViewNet(ds *api.DperService) func(*gin.Context) {
	return func(c *gin.Context) {
		loglogrus.Log.Infof("Http: Get user request -- BackViewNet")
		appG := app.Gin{c}
		data, err := ds.BackViewNet()
		if err != nil {
			loglogrus.Log.Warnf("Http: Reply to user request -- BackViewNet  warn:%s\n", err)
			appG.Response(http.StatusOK, e.ERROR, err)
			return
		} else {
			appG.Response(http.StatusOK, e.SUCCESS, data)
		}
		loglogrus.Log.Infof("Http: Reply to user request -- BackViewNet  succeed!\n")
	}

}

func HelpMenu(ds *api.DperService) func(*gin.Context) {
	return func(c *gin.Context) {
		appG := app.Gin{c}
		data := ds.HelpMenu()
		appG.Response(http.StatusOK, e.SUCCESS, data)
	}
}

func Exit(ds *api.DperService) func(*gin.Context) {
	return func(c *gin.Context) {
		loglogrus.Log.Infof("Http: Get user request -- Exit\n")
		appG := app.Gin{c}
		appG.Response(http.StatusOK, e.SUCCESS, "ByeBye")
		defer ds.Exit()
	}
}

func SignatureReturn(ds *api.DperService) func(*gin.Context) {
	return func(c *gin.Context) {
		appG := app.Gin{c}
		message := c.PostForm("message")
		signature, TxID, err := ds.SignatureReturn(message)
		if err != nil {
			appG.Response(http.StatusOK, e.ERROR, err)
			return
		} else {
			appG.Response(http.StatusOK, e.SUCCESS, fmt.Sprintf("message:%s TxID:%x signature:%x", message, TxID, signature))
		}

	}
}
