package dao

import (
	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/util"
	"encoding/json"
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


	//模块组map:{"1"："knowledge_map"}
	//moduleGroupMap:= make(map[int]interface{}, 0)
	//returnModuleGroupMap:=make(map[string]interface{}, 0)
	//moduleGroup:= make(map[string]interface{}, 0)
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
						println(key)
						b, err := json.Marshal(value)

						if err != nil {
							fmt.Println("json.Marshal failed:", err)
							return nil, err

						}
						println(string(b))
					moduleGroup[key]=value
					}
				}
				resultMap[productCode]=moduleGroup
			}

		}
	}
	return resultMap, nil
}


func (controller MYSQLController) GetRolePrivileges(enterpriseID string,roleId int) (map[string]interface{}, error) {
	//查询产品信息
	queryProductStr := fmt.Sprintf(`SELECT p.cmd_list,m.code,mg.code,m.product,pro.code from %s p
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
	//产品map:{"1"："knowledge"}
	resultMap:= make(map[string]interface{}, 0)
	var resultModuleGroupMap map[string]interface{}
	var pidTemp string
	for rows.Next()  {
		var cmdList,mcode,mgcode,pid,procode string
		err := rows.Scan(&cmdList,&mcode,&mgcode,&pid,&procode)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}
		resultModuleMap:= make(map[string]interface{}, 0)
		cmdListMap:=strings.Split(cmdList, ",")
		resultModuleMap[mcode]=cmdListMap
		if pidTemp!=pid {
			resultModuleGroupMap=make(map[string]interface{}, 0)
			pidTemp=pid
		}
		resultModuleGroupMap[mgcode]=resultModuleMap
		resultMap[procode]=resultModuleGroupMap
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
				if moduleMap[moduleCode]==nil{
					continue
				}
				queryInsertStr := fmt.Sprintf(`INSERT INTO %s (role, module, cmd_list) VALUES (?, ?, ?)`, rolePrivilegeTable)
				_, err = t.Exec(queryInsertStr, roleId, moduleMap[moduleCode].ID, strings.Join(modulePri, ","))
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

