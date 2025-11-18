package api

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vintcessun/XMU-Service/utils"
)

type Zyh365ServicePassword struct {
	client   *resty.Client
	Username string
	Password string
	Token    string
}

type Zyh365LoginResponse struct {
	ErrCode string `json:"errCode"`
	Message string `json:"message"`
	Token   string `json:"token"`
}

func getClientZyh365() *resty.Client {
	ua := utils.GetFakeUAComputer()
	client := resty.New()
	client.SetHeader("User-Agent", ua)
	client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(100))
	return client
}

func (z *Zyh365ServicePassword) Login() error {
	z.client = getClientZyh365()
	resp, err := z.client.R().SetFormData(map[string]string{
		"api":         "userCenter/login",
		"username":    z.Username,
		"password":    z.Password,
		"apimode":     "vmsapi",
		"AccessKeyId": "92ae62f25d4d4ac8a58d58e2476a4e5b",
	}).Post("https://ogmapi.zyh365.com/api/userCenter/login.do")

	if err != nil {
		return err
	}

	loginResp, err := utils.UnmarshalJSON[Zyh365LoginResponse](resp.Body())
	if err != nil {
		return err
	}

	if loginResp.ErrCode != "0000" {
		return fmt.Errorf("登录失败: %s", loginResp.Message)
	}

	z.Token = resp.Header().Get("At")

	return nil
}

type Zyh365ServiceHours struct {
	CreditHours float64 `json:"credit_hours"`
	HonorHours  float64 `json:"honor_hours"`
	TotalHours  float64 `json:"total_hours"`
}

type Zyh365ProfileResponse struct {
	ErrCode            string  `json:"errCode"`
	ExchangeRecordURL  string  `json:"exchange_record_url"`
	HelpURL            string  `json:"help_url"`
	HoursHistory       float64 `json:"hours_history"`
	HoursSystem        float64 `json:"hours_system"`
	HoursTotalDuration float64 `json:"hours_totalduration"`
	Info               struct {
		Avatar           string  `json:"avatar"`
		Birthday         string  `json:"birthday"`
		CardNo           string  `json:"card_no"`
		City             string  `json:"city"`
		CityName         string  `json:"city_name"`
		County           string  `json:"county"`
		CountyName       string  `json:"county_name"`
		CreateTime       int64   `json:"createtime"`
		CredentialsType  int     `json:"credentials_type"`
		CurrentAddress   string  `json:"current_address"`
		Description      string  `json:"description"`
		HoursHistory     float64 `json:"hours_history"`
		HoursSystem      float64 `json:"hours_system"`
		Mobile           string  `json:"mobile"`
		Nation           string  `json:"nation"`
		Nickname         string  `json:"nickname"`
		Points           float64 `json:"points"`
		Political        string  `json:"political"`
		Province         string  `json:"province"`
		ProvinceName     string  `json:"province_name"`
		QQ               string  `json:"qq"`
		Ranking          string  `json:"ranking"`
		RealName         string  `json:"real_name"`
		ServiceDirection string  `json:"service_direction"`
		Sex              string  `json:"sex"`
		StarLevel        int     `json:"star_level"`
		Tel              string  `json:"tel"`
		TrainTime        float64 `json:"trainTime"`
		UserName         string  `json:"username"`
		Vcode            string  `json:"vcode"`
		ZyzStatus        int     `json:"zyz_status"`
		ZyzId            string  `json:"zyzid"`
	} `json:"info"`
	Message string `json:"message"`
	Nav     []struct {
		Name  string                `json:"name"`
		URL   string                `json:"url"`
		Value utils.Float64OrString `json:"value"`
	} `json:"nav"`
	Points    float64 `json:"points"`
	Ranking   string  `json:"ranking"`
	Status    bool    `json:"status"`
	TrainTime float64 `json:"trainTime"`
}

func zyh365NameMatch(name, pattern string) bool {
	// 转换为rune处理中文等多字节字符
	patternRunes := []rune(pattern)
	nameRunes := []rune(name)
	lenPattern := len(patternRunes)
	lenName := len(nameRunes)

	// 步骤1：提取pattern中*左右的固定部分（left：*左边的固定字符，right：*右边的固定字符）
	// 若有多个*，取第一个*左边为left，最后一个*右边为right（多个*等价于单个*，均代表至少1个字符）
	leftEnd := -1    // 第一个*的位置（左侧固定部分结束索引）
	rightStart := -1 // 最后一个*的位置（右侧固定部分开始索引）
	hasStar := false

	for i, c := range patternRunes {
		if c == '*' {
			hasStar = true
			if leftEnd == -1 {
				leftEnd = i // 记录第一个*的位置
			}
			rightStart = i // 不断更新为最后一个*的位置
		}
	}

	// 提取left和right
	var left, right []rune
	if hasStar {
		left = patternRunes[:leftEnd]       // *左边的固定部分（可能为空）
		right = patternRunes[rightStart+1:] // *右边的固定部分（可能为空）
	} else {
		// 无*时，pattern必须与name完全相同
		return lenPattern == lenName && string(patternRunes) == string(nameRunes)
	}

	lenLeft := len(left)
	lenRight := len(right)

	// 步骤2：校验名字长度是否满足最小要求（固定部分长度 + *至少1个字符）
	minRequiredLen := lenLeft + lenRight + 1 // *至少1个字符
	if lenName < minRequiredLen {
		return false
	}

	// 步骤3：校验left是否匹配名字的开头部分
	if lenLeft > 0 {
		// 名字前lenLeft个字符必须与left完全相同
		if lenName < lenLeft {
			return false
		}
		for i := 0; i < lenLeft; i++ {
			if nameRunes[i] != left[i] {
				return false
			}
		}
	}

	// 步骤4：校验right是否匹配名字的结尾部分
	if lenRight > 0 {
		// 名字后lenRight个字符必须与right完全相同
		if lenName < lenRight {
			return false
		}
		nameRightStart := lenName - lenRight // 名字中right开始的索引
		for i := 0; i < lenRight; i++ {
			if nameRunes[nameRightStart+i] != right[i] {
				return false
			}
		}
	}

	// 步骤5：校验left和right在名字中的位置是否合法（left结束位置 < right开始位置，确保中间至少1个字符）
	leftEndInName := lenLeft               // left在名字中结束的索引（下一个字符就是*代表的部分）
	rightStartInName := lenName - lenRight // right在名字中开始的索引
	if leftEndInName >= rightStartInName {
		return false // 中间没有足够字符（*代表至少1个）
	}

	return true
}

func (z *Zyh365ServiceHours) GetHours(tokenStr, name string) error {
	claims := jwt.MapClaims{}
	token, _, err := jwt.NewParser().ParseUnverified(tokenStr, claims)
	if err != nil {
		return err
	}
	userId, err := token.Claims.GetSubject()
	if err != nil {
		return err
	}

	client := getClientZyh365()
	resp, err := client.R().SetFormData(map[string]string{
		"api":     "volunteerinfo/getvolunteerbyId",
		"apimode": "vmsapi",
		"zyzid":   userId,
	}).SetHeader("User-Id", userId).SetHeader("Authorization", tokenStr).SetHeader("Platform-Id", "3").Post("https://ogmapi.zyh365.com/api/volunteerinfo/getvolunteerbyId.do")
	if err != nil {
		return err
	}
	profileResp, err := utils.UnmarshalJSON[Zyh365ProfileResponse](resp.Body())
	if err != nil {
		return err
	}

	if profileResp.ErrCode != "0000" {
		return fmt.Errorf("获取信息失败: %s", profileResp.Message)
	}

	if !zyh365NameMatch(name, profileResp.Info.RealName) || !zyh365NameMatch(name, profileResp.Info.UserName) {
		return fmt.Errorf("姓名不匹配，平台信息片段: %s, %s", profileResp.Info.RealName, profileResp.Info.UserName)
	}

	for _, nav := range profileResp.Nav {
		if nav.Value.IsFloat == false {
			continue
		}
		switch nav.Name {
		case "荣誉时数":
			z.HonorHours = nav.Value.FloatVal
		case "信用时数":
			z.CreditHours = nav.Value.FloatVal
		}
	}

	z.TotalHours = z.HonorHours + z.CreditHours

	return nil
}
