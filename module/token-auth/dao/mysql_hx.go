package dao

import (
	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/util"
	"fmt"
	"strings"
)

func (controller MYSQLController) GetRolesHX(enterpriseID string) ([]*data.RoleHX, error) {
	queryStr := fmt.Sprintf("SELECT id, uuid, name, description FROM %s WHERE enterprise = ?", roleTable)
	roleRows, err := controller.connectDB.Query(queryStr,enterpriseID)
	roles := make([]*data.RoleHX, 0)
	if err != nil {
		return nil, err
	}
	defer roleRows.Close()

	for roleRows.Next() {
		role := data.RoleHX{}
		err := roleRows.Scan(&role.ID,&role.UUID, &role.Name, &role.Description)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}
		roles = append(roles, &role)
	}
	return roles, nil
}

func (controller MYSQLController) GetModulesHX() (map[string]interface{}, error) {
	//查询产品信息
	queryProductStr := fmt.Sprintf("SELECT id, code FROM %s ", productTabel)
	productRows, productErr := controller.connectDB.Query(queryProductStr)
	if productErr != nil {
		return nil, productErr
	}
	defer productRows.Close()


	queryModuleGroupStr := fmt.Sprintf("SELECT id, code,product FROM %s ", moduleGroup)
	moduleGrouprows, moduleGrouErr := controller.connectDB.Query(queryModuleGroupStr)
	if moduleGrouErr != nil {
		return nil, moduleGrouErr
	}
	defer moduleGrouprows.Close()


	queryModulesStr := fmt.Sprintf("SELECT code,product,m.group,cmd_list FROM %s m", moduleTable)
	modulesRows, moduleErr := controller.connectDB.Query(queryModulesStr)
	if moduleErr != nil {
		return nil, moduleErr
	}
	defer modulesRows.Close()
	resultMap:= make(map[string]interface{}, 0)

	moduleMapArray:= make(map[int][]map[string][]string, 0)
	for modulesRows.Next()  {
		var product,group int
		var code,cmdList string
		err := modulesRows.Scan(&code,&product,&group,&cmdList)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}
		cmdListMap:=strings.Split(cmdList, ",")
		moduleMap:= make(map[string][]string, 0)
		moduleMap[code]=cmdListMap
		moduleMapArray[group] = append(moduleMapArray[group],moduleMap )

	}

	moduleGroupMapArray:= make(map[int][]map[string]map[string][]string, 0)
	for moduleGrouprows.Next()  {
		moduleGroup:= make(map[string]map[string][]string, 0)
		var product,id int
		var code string
		err := moduleGrouprows.Scan(&id,&code,&product)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}
		for group, modulePri := range moduleMapArray {
			if(group==id) {
				moduleMap:= make(map[string][]string, 0)
				for _, moduleValue := range modulePri {
					for key, value := range moduleValue {
						moduleMap[key]=value
					}
				}
				moduleGroup[code]=moduleMap
			}
		}
		moduleGroupMapArray[product] = append(moduleGroupMapArray[product],moduleGroup)

	}

	//产品map:{"1"："knowledge"}
	productMap:=make(map[string]int, 0)
	for productRows.Next()  {
		var id int
		var code string
		err := productRows.Scan(&id,&code)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}
		productMap[code]=id
	}
	for productCode, productId := range productMap {
		for moduleProduct, modulePri := range moduleGroupMapArray {
			if(moduleProduct==productId){
				moduleGroup:= make(map[string]map[string][]string, 0)
				for _, groupValue := range modulePri {
					for key, value := range groupValue {
						moduleGroup[key]=value
					}
				}
				resultMap[productCode]=moduleGroup
			}

		}
	}
	return resultMap, nil
}


func (controller MYSQLController) GetRolePrivileges(enterpriseID string,roleId int) (map[string]map[string]map[string][]string, error) {
	//查询产品信息
	queryProductStr := fmt.Sprintf(`SELECT p.cmd_list,m.code,mg.code,m.product,pro.code,m.group from %s p
	LEFT JOIN %s m on p.module=m.id
	LEFT JOIN %s mg on m.group=mg.id
	LEFT JOIN %s pro on m.product=pro.id
	where p.role = ? and m.enterprise=?
	`, rolePrivilegeTable,moduleTable,moduleGroup,productTabel)
	rows, productErr := controller.connectDB.Query(queryProductStr,roleId,enterpriseID)
	if productErr != nil {
		return nil, productErr
	}
	defer rows.Close()
	//产品map
	resultMap:= make(map[string]map[string]map[string][]string, 0)
	for rows.Next()  {
		var cmdList,mcode,mgcode,procode string
		var pid,group int
		err := rows.Scan(&cmdList,&mcode,&mgcode,&pid,&procode,&group)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}
		resultModuleMap:= make(map[string][]string, 0)
		cmdListMap:=strings.Split(cmdList, ",")
		resultModuleMap[mcode]=cmdListMap
		if resultMap[procode]==nil {//新产品
			resultMap[procode]=make(map[string]map[string][]string, 0)
			if resultMap[procode][mgcode]==nil {
				resultMap[procode][mgcode]=make(map[string][]string, 0)
				resultMap[procode][mgcode]=resultModuleMap
			}else{
				resultMap[procode][mgcode][mcode]=cmdListMap
			}
		}else{//已存在产品
			if resultMap[procode][mgcode]==nil {
				resultMap[procode][mgcode]=make(map[string][]string, 0)
				resultMap[procode][mgcode]=resultModuleMap
			}else{
				resultMap[procode][mgcode][mcode]=cmdListMap
			}
		}
	}
	return resultMap, nil
}

func (controller MYSQLController) GetUserPrivileges(enterpriseID string,userCode string) (map[string]map[string]map[string][]string, error) {
	//查询产品信息
	queryProductStr :=`SELECT p.cmd_list,m.code,mg.code,m.product,pro.code,m.group from privileges p
	LEFT JOIN  modules m on p.module=m.id
	LEFT JOIN  module_group mg on m.group=mg.id
	LEFT JOIN  product pro on m.product=pro.id
	where p.role in (select id from roles where uuid in(select role from user_privileges where human=?) )
	and m.enterprise=?
	`
	rows, productErr := controller.connectDB.Query(queryProductStr,userCode,enterpriseID)
	if productErr != nil {
		return nil, productErr
	}
	defer rows.Close()
	//产品map
	resultMap:= make(map[string]map[string]map[string][]string, 0)
	for rows.Next()  {
		var cmdList,mcode,mgcode,procode string
		var pid,group int
		err := rows.Scan(&cmdList,&mcode,&mgcode,&pid,&procode,&group)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}
		resultModuleMap:= make(map[string][]string, 0)
		cmdListMap:=strings.Split(cmdList, ",")
		resultModuleMap[mcode]=cmdListMap
		if resultMap[procode]==nil {//新产品
			resultMap[procode]=make(map[string]map[string][]string, 0)
			if resultMap[procode][mgcode]==nil {
				resultMap[procode][mgcode]=make(map[string][]string, 0)
				resultMap[procode][mgcode]=resultModuleMap
			}else{
				resultMap[procode][mgcode][mcode]=cmdListMap
			}
		}else{//已存在产品
			if resultMap[procode][mgcode]==nil {
				resultMap[procode][mgcode]=make(map[string][]string, 0)
				resultMap[procode][mgcode]=resultModuleMap
			}else{
				resultMap[procode][mgcode][mcode]=cmdListMap
			}
		}
	}
	return resultMap, nil
}

func (controller MYSQLController) UpdateRolePrivileges(enterpriseID string,roleId int,privileges map[string]map[string]map[string][]string) error {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return err
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		util.LogDBError(err)
		return err
	}
	defer util.ClearTransition(t)

	//删除权限
	var queryStr = fmt.Sprintf(`DELETE FROM %s WHERE role = ?`, rolePrivilegeTable)
	_, err = t.Exec(queryStr, roleId)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	//查询功能模块
	moduleMap := map[string]*data.ModuleHX{}
	queryModulesStr := fmt.Sprintf("SELECT id,code,product,m.group,cmd_list FROM %s m", moduleTable)
	modulesRows, moduleErr := controller.connectDB.Query(queryModulesStr)
	if moduleErr != nil {
		return moduleErr
	}
	defer modulesRows.Close()


	for modulesRows.Next()  {
		module:=data.ModuleHX{}
		err := modulesRows.Scan(&module.ID,&module.Code,&module.Product,&module.Group,&module.CmdList)
		if err != nil {
			util.LogDBError(err)
			return err
		}
		moduleMap[module.Code] = &module
	}

	for _, product := range privileges {
		for _, moduleGroup := range product {
			for moduleCode, modulePri := range moduleGroup {
				if moduleCode==""|| moduleMap[moduleCode]==nil{
					continue
				}
				queryInsertStr := fmt.Sprintf(`INSERT INTO %s (role, module, cmd_list) VALUES (?, ?, ?)`, rolePrivilegeTable)
				var cmdList string
				var module int
				if modulePri[0]!=""{
					cmdList = strings.Join(modulePri, ",")
				}
				module = moduleMap[moduleCode].ID
				_, err = t.Exec(queryInsertStr, roleId,module , cmdList)
				if err != nil {
					util.LogDBError(err)
					return err
				}
			}

		}

	}
	err = t.Commit()
	if err != nil {
		util.LogDBError(err)
		return err
	}
	return nil
}

func (controller MYSQLController) GetLabelUsers(enterpriseID string ,role int)([]*data.UserHX, error) {
	//IN ('my_mission_marked ',' my_auditing_tasks ',' my_split_tasks ')
	queryStr := fmt.Sprintf(`SELECT id,user_name FROM users WHERE uuid IN (
SELECT human FROM user_privileges WHERE role IN (
SELECT uuid FROM roles WHERE enterprise=? AND id IN (
SELECT role FROM privileges WHERE module IN (
SELECT id FROM modules WHERE CODE = ?))))`)
	code := ""
	if role ==1 {//标注
		code="my_mission_marked"
	}else if role ==2{//审核
		code="my_auditing_tasks"
	}else if role ==3{//切分
		code="my_split_tasks"
	}
	userRows, err := controller.connectDB.Query(queryStr,enterpriseID,code)
	users := make([]*data.UserHX, 0)
	if err != nil {
		return nil, err
	}
	defer userRows.Close()

	for userRows.Next() {
		user := data.UserHX{}
		err := userRows.Scan(&user.ID,&user.UserName)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

func (controller MYSQLController) GetUserAccessInfo(
	userCode string) (*data.UserHX, error) {
	user := data.UserHX{}
	queryStr := fmt.Sprintf(`
		SELECT id, user_name, password
		FROM %s
		WHERE uuid = ?`, userTableV3)

	userRows, err := controller.connectDB.Query(queryStr,userCode)
	for userRows.Next() {
		err = userRows.Scan(&user.ID, &user.UserName, &user.Password)
		if err != nil {
			util.LogDBError(err)
			return nil, nil
		}
	}
	return &user, nil
}