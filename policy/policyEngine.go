// policyEngine.go
package policy

import (
	 "utils/patriciaDB"
	  "utils/policy/policyCommonDefs"
	  "utils/netUtils"
	  "strings"
	 "reflect"
	 "sort"
	 "strconv"
//	"utils/commonDefs"
//	"net"
//	"asicdServices"
//	"asicd/asicdConstDefs"
	"bytes"
  //  "database/sql"
    "fmt"
)
func (db *PolicyEngineDB) ActionListHasAction(actionList []string, actionType int, action string) (match bool) {
	fmt.Println("ActionListHasAction for action ", action)
	return match
}
func (db *PolicyEngineDB) PolicyEngineCheck(route interface{}, policyType int) (actionList []string){
	fmt.Println("PolicyEngineTest to see if there are any policies  ")
	return nil
}
func (db *PolicyEngineDB) PolicyEngineImplementActions(entity PolicyEngineFilterEntityParams, policyStmt PolicyStmt, params interface {}) (actionList []string){
	fmt.Println("policyEngineImplementActions")
	if policyStmt.Actions == nil {
		fmt.Println("No actions")
		return actionList
	}
	var i int
	addActionToList := false
	for i=0;i<len(policyStmt.Actions);i++ {
	  addActionToList = false
	  fmt.Printf("Find policy action number %d name %s in the action database\n", i, policyStmt.Actions[i])
	  actionItem := db.PolicyActionsDB.Get(patriciaDB.Prefix(policyStmt.Actions[i]))
	  if actionItem == nil {
	     fmt.Println("Did not find action ", policyStmt.Actions[i], " in the action database")	
		 continue
	  }
	  action := actionItem.(PolicyAction)
	  fmt.Printf("policy action number %d type %d\n", i, action.ActionType)
		switch action.ActionType {
		   case policyCommonDefs.PolicyActionTypeRouteDisposition:
		      fmt.Println("PolicyActionTypeRouteDisposition action to be applied")
	           addActionToList = true
			  if db.ActionfuncMap[policyCommonDefs.PolicyActionTypeRouteDisposition] != nil {
			     db.ActionfuncMap[policyCommonDefs.PolicyActionTypeRouteDisposition](action.ActionInfo,params)	
			  }
			  break
		   case policyCommonDefs.PolicyActionTypeRouteRedistribute:
		      fmt.Println("PolicyActionTypeRouteRedistribute action to be applied")
			  if db.ActionfuncMap[policyCommonDefs.PolicyActionTypeRouteRedistribute] != nil {
			     db.ActionfuncMap[policyCommonDefs.PolicyActionTypeRouteRedistribute](action.ActionInfo,params)	
			  }
	          addActionToList = true
			  break
		   default:
		      fmt.Println("UnknownInvalid type of action")
			  break
		}
		if addActionToList == true {
		   if actionList == nil {
		      actionList = make([]string,0)
		   }
	       actionList = append(actionList,action.Name)
		}
	}
    return actionList
}
func (db *PolicyEngineDB) FindPrefixMatch(ipAddr string, ipPrefix patriciaDB.Prefix, policyName string)(match bool){
    fmt.Println("Prefix match policy ", policyName)
	policyListItem := db.PrefixPolicyListDB.GetLongestPrefixNode(ipPrefix)
	if policyListItem == nil {
		fmt.Println("intf stored at prefix ", ipPrefix, " is nil")
		return false
	}
    if policyListItem != nil && reflect.TypeOf(policyListItem).Kind() != reflect.Slice {
		fmt.Println("Incorrect data type for this prefix ")
		 return false
	}
	policyListSlice := reflect.ValueOf(policyListItem)
	for idx :=0;idx < policyListSlice.Len();idx++ {
	   prefixPolicyListInfo := policyListSlice.Index(idx).Interface().(PrefixPolicyListInfo)
	   if prefixPolicyListInfo.policyName != policyName {
	      fmt.Println("Found a potential match for this prefix but the policy ", policyName, " is not what we are looking for")
		  continue
	   }
	   if prefixPolicyListInfo.lowRange == -1 && prefixPolicyListInfo.highRange == -1 {
          fmt.Println("Looking for exact match condition for prefix ", prefixPolicyListInfo.ipPrefix)
		  if bytes.Equal(ipPrefix, prefixPolicyListInfo.ipPrefix) {
			 fmt.Println(" Matched the prefix")
	         return true
		  }	else {
			 fmt.Println(" Did not match the exact prefix")
		     return false	
		  }
	   }
	   tempSlice:=strings.Split(ipAddr,"/")
	   maskLen,err:= strconv.Atoi(tempSlice[1])
	   if err != nil {
	       fmt.Println("err getting maskLen")
		   return false	
	   }
	   fmt.Println("Mask len = ", maskLen)
	   if maskLen < prefixPolicyListInfo.lowRange || maskLen > prefixPolicyListInfo.highRange {
	      fmt.Println("Mask range of the route ", maskLen , " not within the required mask range:", prefixPolicyListInfo.lowRange,"..", prefixPolicyListInfo.highRange)	
		  return false
	   } else {
	      fmt.Println("Mask range of the route ", maskLen , " within the required mask range:", prefixPolicyListInfo.lowRange,"..", prefixPolicyListInfo.highRange)	
		  return true
	   }
	} 
	return match
}
func (db *PolicyEngineDB) PolicyEngineMatchConditions(entity PolicyEngineFilterEntityParams, policyStmt PolicyStmt) (match bool, conditionsList []string){
    fmt.Println("policyEngineMatchConditions")
	var i int
	allConditionsMatch := true
	anyConditionsMatch := false
	addConditiontoList := false
	for i=0;i<len(policyStmt.Conditions);i++ {
	  addConditiontoList = false
	  fmt.Printf("Find policy condition number %d name %s in the condition database\n", i, policyStmt.Conditions[i])
	  conditionItem := db.PolicyConditionsDB.Get(patriciaDB.Prefix(policyStmt.Conditions[i]))
	  if conditionItem == nil {
	     fmt.Println("Did not find condition ", policyStmt.Conditions[i], " in the condition database")	
		 continue
	  }
	  condition := conditionItem.(PolicyCondition)
	  fmt.Printf("policy condition number %d type %d\n", i, condition.ConditionType)
      switch condition.ConditionType {
		case policyCommonDefs.PolicyConditionTypeDstIpPrefixMatch:
		  fmt.Println("PolicyConditionTypeDstIpPrefixMatch case")
		  ipPrefix,err := netUtils.GetNetworkPrefixFromCIDR(entity.DestNetIp)
		  if err != nil {
			fmt.Println("Invalid ipPrefix for the route ", entity.DestNetIp)
			return match,conditionsList
		  }
		  match := db.FindPrefixMatch(entity.DestNetIp, ipPrefix,policyStmt.Name)
		  if match {
		    fmt.Println("Found a match for this prefix")
			anyConditionsMatch = true
			addConditiontoList = true
		  }
		break
		case policyCommonDefs.PolicyConditionTypeProtocolMatch:
		  fmt.Println("PolicyConditionTypeProtocolMatch case")
		  matchProto := condition.ConditionInfo.(string)
		  if matchProto == entity.RouteProtocol {
			fmt.Println("Protocol condition matches")
			anyConditionsMatch = true
			addConditiontoList = true
		  } 
		break
		default:
		  fmt.Println("Not a known condition type")
          break
	  }
	  if addConditiontoList == true{
		if conditionsList == nil {
		   conditionsList = make([]string,0)
		}
		conditionsList = append(conditionsList,condition.Name)
	  }
	}
   if policyStmt.MatchConditions == "all" && allConditionsMatch == true {
	return true,conditionsList
   }
   if policyStmt.MatchConditions == "any" && anyConditionsMatch == true {
	return true,conditionsList
   }
    return match,conditionsList
}
func (db *PolicyEngineDB) PolicyEngineApplyPolicyStmt(entity *PolicyEngineFilterEntityParams, policy Policy, policyStmt PolicyStmt, policyPath int, params interface{}, hit *bool, deleted *bool) {
	fmt.Println("policyEngineApplyPolicyStmt - ", policyStmt.Name)
	var conditionList []string
	if policyStmt.Conditions == nil {
		fmt.Println("No policy conditions")
		*hit=true
	} else {
	   match,ret_conditionList := db.PolicyEngineMatchConditions(*entity, policyStmt)
	   fmt.Println("match = ", match)
	   *hit = match
	   if !match {
		   fmt.Println("Conditions do not match")
		   return
	   }
	   if ret_conditionList != nil {
		 if conditionList == nil {
			conditionList = make([]string,0)
		 }
		 for j:=0;j<len(ret_conditionList);j++ {
			conditionList =append(conditionList,ret_conditionList[j])
		 }
	   }
	}
	actionList := db.PolicyEngineImplementActions(*entity, policyStmt, params)
	if db.ActionListHasAction(actionList, policyCommonDefs.PolicyActionTypeRouteDisposition,"Reject") {
		fmt.Println("Reject action was applied for this entity")
		*deleted = true
	}
	//check if the route still exists - it may have been deleted by the previous statement action
	if db.IsEntityPresentFunc != nil {
		*deleted = !(db.IsEntityPresentFunc(params))
	}
	if db.UpdateEntityDB != nil {
		policyDetails := PolicyDetails{Policy:policy.Name, PolicyStmt:policyStmt.Name,ConditionList:conditionList,ActionList:actionList, EntityDeleted:*deleted}
		db.UpdateEntityDB(policyDetails,params)
	}
}

func (db *PolicyEngineDB) PolicyEngineApplyPolicy(entity *PolicyEngineFilterEntityParams, policy Policy, policyPath int,params interface{}, hit *bool) {
	fmt.Println("policyEngineApplyPolicy - ", policy.Name)
     var policyStmtKeys []int
	 deleted := false
	 for k:=range policy.PolicyStmtPrecedenceMap {
		fmt.Println("key k = ", k)
		policyStmtKeys = append(policyStmtKeys,k)
	}
	sort.Ints(policyStmtKeys)
	for i:=0;i<len(policyStmtKeys);i++ {
		fmt.Println("Key: ", policyStmtKeys[i], " policyStmtName ", policy.PolicyStmtPrecedenceMap[policyStmtKeys[i]])
		policyStmt := db.PolicyStmtDB.Get((patriciaDB.Prefix(policy.PolicyStmtPrecedenceMap[policyStmtKeys[i]])))
        if policyStmt == nil {
			fmt.Println("Invalid policyStmt")
			continue
		}
		db.PolicyEngineApplyPolicyStmt(entity,policy,policyStmt.(PolicyStmt),policyPath, params, hit, &deleted)
		if deleted == true {
			fmt.Println("Entity was deleted as a part of the policyStmt ", policy.PolicyStmtPrecedenceMap[policyStmtKeys[i]])
             break
		}
		if *hit == true {
			if policy.MatchType == "any" {
				fmt.Println("Match type for policy ", policy.Name, " is any and the policy stmt ", (policyStmt.(PolicyStmt)).Name, " is a hit, no more policy statements will be executed")
				break
			}
		}
	}
}

func (db *PolicyEngineDB) PolicyEngineTraverseAndApplyPolicy(policy Policy) {
	fmt.Println("PolicyEngineTraverseAndApplyPolicy -  apply policy ", policy.Name)
    if policy.ExportPolicy || policy.ImportPolicy{
	   fmt.Println("Applying import/export policy to all routes")
	  // PolicyEngineTraverseAndApply(policy)
	} else if policy.GlobalPolicy {
		fmt.Println("Need to apply global policy")
		//policyEngineApplyGlobalPolicy(policy)
	}
}

func (db *PolicyEngineDB) PolicyEngineTraverseAndReversePolicy(policy Policy){
	fmt.Println("PolicyEngineTraverseAndReversePolicy -  reverse policy ", policy.Name)
    if policy.ExportPolicy || policy.ImportPolicy{
	   fmt.Println("Reversing import/export policy ")
	   //PolicyEngineTraverseAndReverse(policy)
	} else if policy.GlobalPolicy {
		fmt.Println("Need to reverse global policy")
		//policyEngineReverseGlobalPolicy(policy)
	}
	
}
func (db *PolicyEngineDB) PolicyEngineFilter(entity PolicyEngineFilterEntityParams, policyPath int, params interface{}) {
	fmt.Println("PolicyEngineFilter")
	var policyPath_Str string
	if policyPath == policyCommonDefs.PolicyPath_Import {
	   policyPath_Str = "Import"
	} else if policyPath == policyCommonDefs.PolicyPath_Export {
	   policyPath_Str = "Export"
	} else if policyPath == policyCommonDefs.PolicyPath_All {
		policyPath_Str = "ALL"
		fmt.Println("policy path ", policyPath_Str, " unexpected in this function")
		return
	}
	fmt.Println("PolicyEngineFilter for policypath ", policyPath_Str, "create = ", entity.CreatePath, " delete = ", entity.DeletePath, " route: ", entity.DestNetIp, " protocol type: ", entity.RouteProtocol)
    var policyKeys []int
	var policyHit bool
	idx :=0
	var policyInfo interface{}
	if policyPath == policyCommonDefs.PolicyPath_Import{
	   for k:=range db.ImportPolicyPrecedenceMap {
	      policyKeys = append(policyKeys,k)
	   }
	} else if policyPath == policyCommonDefs.PolicyPath_Export{
	   for k:=range db.ExportPolicyPrecedenceMap {
	      policyKeys = append(policyKeys,k)
	   }
	}
	sort.Ints(policyKeys)
	for ;; {
		if entity.DeletePath == true {		//policyEngineFilter called during delete
			if entity.PolicyList != nil {
             if idx >= len(entity.PolicyList) {
				break
			 } 		
		     fmt.Println("getting policy ", idx, " from entity.PolicyList")
	         policyInfo = 	db.PolicyDB.Get(patriciaDB.Prefix(entity.PolicyList[idx]))
		     idx++
			 if policyInfo.(Policy).ExportPolicy && policyPath == policyCommonDefs.PolicyPath_Import || policyInfo.(Policy).ImportPolicy && policyPath == policyCommonDefs.PolicyPath_Export {
				fmt.Println("policy ", policyInfo.(Policy).Name, " not the same type as the policypath -", policyPath_Str)
				continue
			 } 
	        } else {
		      fmt.Println("PolicyList empty and this is a delete operation, so break")
               break
	        }		
	    }  else if entity.CreatePath == true{ //policyEngine filter called during create 
			fmt.Println("idx = ", idx, " len(policyKeys):", len(policyKeys))
            if idx >= len(policyKeys) {
				break
			}		
			policyName := ""
            if policyPath == policyCommonDefs.PolicyPath_Import {
               policyName = db.ImportPolicyPrecedenceMap[policyKeys[idx]]
			} else if policyPath == policyCommonDefs.PolicyPath_Export {
               policyName = db.ExportPolicyPrecedenceMap[policyKeys[idx]]
			}
		    fmt.Println("getting policy  ", idx, " policyKeys[idx] = ", policyKeys[idx]," ", policyName," from PolicyDB")
             policyInfo = db.PolicyDB.Get((patriciaDB.Prefix(policyName)))
			idx++
	      }
	      if policyInfo == nil {
	        fmt.Println("Nil policy")
		    continue
	      }
	      policy := policyInfo.(Policy)
		  localPolicyDB := *db.LocalPolicyDB
	      if localPolicyDB != nil && localPolicyDB[policy.LocalDBSliceIdx].IsValid == false {
	        fmt.Println("Invalid policy at localDB slice idx ", policy.LocalDBSliceIdx)
		    continue	
	      }		
	      db.PolicyEngineApplyPolicy(&entity, policy, policyPath, params, &policyHit)
	      if policyHit {
	         fmt.Println("Policy ", policy.Name, " applied to the route")	
		     break
	      }
	}
	if entity.PolicyHitCounter == 0{
		fmt.Println("Need to apply default policy, policyPath = ", policyPath, "policyPath_Str= ", policyPath_Str)
		if policyPath == policyCommonDefs.PolicyPath_Import {
		   fmt.Println("Applying default import policy")
			if db.DefaultImportPolicyActionFunc != nil {
				db.DefaultImportPolicyActionFunc(nil,params)
			}
		} else if policyPath == policyCommonDefs.PolicyPath_Export {
			fmt.Println("Applying default export policy")
			if db.DefaultExportPolicyActionFunc != nil {
				db.DefaultExportPolicyActionFunc(nil,params)
			}
		}
	}
}


