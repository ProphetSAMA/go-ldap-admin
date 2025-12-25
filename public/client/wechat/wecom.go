package wechat

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/eryajf/go-ldap-admin/config"
	"github.com/eryajf/go-ldap-admin/public/tools"
	"github.com/wenerme/go-wecom/wecom"
)

// 公共方法：根据中文姓名与手机号生成用户名（拼音 + 手机后四位），手机号不足或缺失则仅拼音
func BuildUsernameWithMobileSuffix(chineseName, mobile string) string {
	base := tools.ConvertToPinYin(chineseName)
	if mobile == "" {
		return base
	}
	if len(mobile) < 4 {
		return base
	}
	return base + mobile[len(mobile)-4:]
}

// 公共方法：清洗邮箱，仅保留本地部分的字母、数字、点、下划线、短横线
// 返回清洗后的邮箱；若无@或域为空返回原值
func SanitizeEmail(email string) string {
	if email == "" || !strings.Contains(email, "@") {
		return email
	}
	parts := strings.SplitN(email, "@", 2)
	local, domain := parts[0], parts[1]
	re := regexp.MustCompile(`[^A-Za-z0-9._-]`)
	cleanLocal := re.ReplaceAllString(local, "")
	if cleanLocal == "" || domain == "" {
		return email
	}
	return cleanLocal + "@" + domain
}

// 官方文档： https://developer.work.weixin.qq.com/document/path/90208
// GetAllDepts 获取所有部门
func GetAllDepts() (ret []map[string]any, err error) {
	depts, err := InitWeComClient().ListDepartment(
		&wecom.ListDepartmentRequest{},
	)
	if err != nil {
		return nil, err
	}
	for _, dept := range depts.Department {
		ele := make(map[string]any)
		ele["name"] = dept.Name
		ele["custom_name_pinyin"] = tools.ConvertToPinYin(dept.Name)
		ele["id"] = dept.ID
		ele["name_en"] = dept.NameEn
		ele["parentid"] = dept.ParentID
		ret = append(ret, ele)
	}
	return ret, nil
}

// 官方文档： https://developer.work.weixin.qq.com/document/path/90201
// GetAllUsers 获取所有员工信息
func GetAllUsers() (ret []map[string]any, err error) {
	depts, err := GetAllDepts()
	if err != nil {
		return nil, err
	}
	for _, dept := range depts {
		users, err := InitWeComClient().ListUser(
			&wecom.ListUserRequest{
				DepartmentID: fmt.Sprintf("%d", dept["id"].(int)),
				FetchChild:   "1",
			},
		)
		if err != nil {
			return nil, err
		}
		for _, user := range users.UserList {
			ele := make(map[string]any)
			ele["name"] = user.Name
			ele["custom_name_pinyin"] = tools.ConvertToPinYin(user.Name)

			// 使用公共方法生成 custom_username
			ele["custom_username"] = BuildUsernameWithMobileSuffix(user.Name, user.Mobile)

			ele["userid"] = user.UserID
			ele["mobile"] = user.Mobile
			ele["position"] = user.Position
			ele["gender"] = user.Gender
			ele["email"] = user.Email
			if user.Email != "" {
				ele["custom_nickname_email"] = strings.Split(user.Email, "@")[0]
			}
			// 使用公共方法生成 mail_sanitized
			if user.Email != "" {
				ele["mail_sanitized"] = SanitizeEmail(user.Email)
			}

			ele["biz_email"] = user.BizMail
			if user.BizMail != "" {
				ele["custom_nickname_biz_email"] = strings.Split(user.BizMail, "@")[0]
			}
			ele["avatar"] = user.Avatar
			ele["telephone"] = user.Telephone
			ele["alias"] = user.Alias
			ele["external_position"] = user.ExternalPosition
			ele["address"] = user.Address
			ele["open_userid"] = user.OpenUserID
			ele["main_department"] = user.MainDepartment
			ele["english_name"] = user.EnglishName
			// 部门ids
			var sourceDeptIds []string
			for _, deptId := range user.Department {
				sourceDeptIds = append(sourceDeptIds, fmt.Sprintf("%s_%d", config.Conf.WeCom.Flag, deptId))
			}
			ele["department_ids"] = sourceDeptIds
			ret = append(ret, ele)
		}
	}
	return ret, nil
}
