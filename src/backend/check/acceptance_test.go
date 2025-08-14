//go:build acceptance

package check

/* TODO uncomment and fix; problem is most likely that the TEST profile is used during that test, not PROD profile
func TestCorsHeaderNotPresentInProd(t *testing.T) {
	hub := GetHub()
	defer hub.WipeData()
	response, err := hub.Parent.DoRequestWithFullResponse("/api/apps/list", nil)
	assert.Nil(t, err)

	assert.Equal(t, "", response.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "", response.Header.Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "", response.Header.Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "", response.Header.Get("Access-Control-Allow-Headers"))
}


func TestDontAllowDifferentHostAndOriginHeader(t *testing.T) {
	hub := GetHub()
	hub.Parent.Origin = "localhost2"
	_, err := hub.Parent.DoRequestWithFullResponse("/api/apps/list", nil)
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(400, "When 'Origin' header is set, it must match host header"), err.Error())
}
*/
