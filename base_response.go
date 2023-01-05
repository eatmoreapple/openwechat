package openwechat

type Ret int

const (
	ticketError         Ret = -14  // ticket error
	logicError          Ret = -2   // logic error
	sysError            Ret = -1   // sys error
	paramError          Ret = 1    // param error
	failedLoginWarn     Ret = 1100 // failed login warn
	failedLoginCheck    Ret = 1101 // failed login check
	cookieInvalid       Ret = 1102 // cookie invalid
	loginEnvAbnormality Ret = 1203 // login environmental abnormality
	optTooOften         Ret = 1205 // operate too often
)

// BaseResponse 大部分返回对象都携带该信息
type BaseResponse struct {
	Ret    Ret
	ErrMsg string
}

func (b BaseResponse) Ok() bool {
	return b.Ret == 0
}

func (b BaseResponse) Err() error {
	if b.Ok() {
		return nil
	}
	return b.Ret
}
