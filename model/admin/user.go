package admin

import (
	"adiris/pkg/config"
	"adiris/pkg/e"
	"adiris/tools/common"
	"adiris/tools/mysql"
)

type User struct {
	Id       string `xorm:"varchar(100) notnull pk index(account) unique 'id'"` //主键id
	IsDel    int
	Remark   string
	Status   int
	Account  string `xorm:"varchar(100) notnull unique 'account'"` //账号
	Password string //密码
	Email    string //邮箱
	Phone    string //手机号码
	at       `xorm:"extends"`
}

/**
用户登录
*/

func Login(account, pwd, gid string) int {
	uid := commons.EncodeMd5(commons.StringJoin(account, gid))
	pw := commons.EncodeMd5(commons.StringJoin(pwd, account, gid))
	u := new(User)
	if !CheckBool(config.EngDb.Where("id = ? and password = ? and is_del = 1 and status = 1 ", uid, pw).Get(&u)) {
		return e.UserNotExist
	}

	/**
	用户存在,这时候查询权限同步权限
	*/

	return e.Success

}

/**
查询用户权限
*/
func GetUserPer() (ret []map[string]string) {
	sql := " SELECT m.`url`, m.`name`, m.`type` FROM " +
		"`user` u, user_role ur, role_menu rm, `menu` m " +
		" WHERE u.`id` = ur.`uid` AND ur.`rid` = rm.`rid` " +
		" AND rm.`mid` = m.`id` " +
		" AND u.`id` = '899d2e464b258017d51af34043bda85a' " +
		"UNION " +
		" SELECT m.`url`, m.`name`, m.`type` " +
		" FROM `user` u, `user_menu` um, `menu` m " +
		" WHERE u.`id` = um.uid AND um.mid = m.id AND u.`id` = '899d2e464b258017d51af34043bda85a'"
	ret, _ = config.EngDb.QueryString(sql)
	return ret
}

/**
用户注册:
account string   用户名
pwd  string   密码
gid  string   机构id(组id)
*/

func Register(account, pwd, gid string) int {
	uid := commons.EncodeMd5(commons.StringJoin(account, gid))
	pw := commons.EncodeMd5(commons.StringJoin(pwd, account, gid))

	mu := make(chan map[string]interface{})
	mg := make(chan map[string]interface{})

	//if GetUserByUid(uid) || !GetGidExist(gid) {
	//	return e.UserExist
	//}

	//if GetUserByUidTb(uid) {
	//	return e.UserExist
	//}
	//if !GetGidExistTb(gid) {
	//	return e.GroupNotExist
	//}
	go GetUserByUid(uid, mu)
	go GetGidExist(gid, mg)
	rmu := <-mu
	rmg := <-mg
	if rmu["flag"] == true || rmg["flag"] == false {
		return e.UserExist
	}

	close(mu)
	close(mg)

	u := User{Account: account, Password: pw, Id: uid, IsDel: 1, Status: 0}
	ug := UserGroup{Uid: uid, Gid: gid, Id: commons.EncodeMd5(commons.StringJoin(uid, gid)), IsDel: 1, Status: 0}
	//给一个默认的角色id
	ur := UserRole{Uid: uid, Rid: "1111", Id: commons.EncodeMd5(commons.StringJoin(uid, "111")), IsDel: 1, Status: 0}

	s := config.EngDb.NewSession()
	var err error
	//这里是不是session多关闭了一次
	defer s.Clone()
	if err = s.Begin(); err != nil {
		return e.Error
	}

	defer s.Rollback()
	if !(mysql.CheckErrInsert(s, u) && mysql.CheckErrInsert(s, ug) && mysql.CheckErrInsert(s, ur)) {
		return e.UserRegisterFail
	}
	s.Commit()
	s.Close()

	return e.Success
}

/**
查找用户是否存在 true：存在 false 不存在
*/
func GetUserByUid(uid string, c chan map[string]interface{}) {
	exist, err := config.EngDb.Where("id = ? ", uid).Exist(&User{})

	m := make(map[string]interface{})
	if commons.CheckErr(err, exist) && exist {
		m["flag"] = true
		c <- m
	} else {
		m["flag"] = false
		c <- m
	}
}

func GetUserByUidTb(uid string) bool {
	exist, err := config.EngDb.Where("id = ? ", uid).Exist(&User{})

	//m := make(map[string]interface{})
	if commons.CheckErr(err, exist) && exist {
		return true
	} else {
		return false
	}
}
