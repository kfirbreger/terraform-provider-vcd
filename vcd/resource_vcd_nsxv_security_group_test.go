package vcd

import (
	//"fmt"
	"testing"
	//"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	//"github.com/hashicorp/terraform-plugin-sdk/terraform"
	//"github.com/vmware/go-vcloud-director/v2/govcd"
)

/*
func getTestConfig() StringMap {
    return StringMap{
        "Org":              testConfig.VCD.Org,
		"Vdc":              testConfig.VCD.Vdc,
		"EdgeGateway":      testConfig.Networking.EdgeGateway,
		"ExternalIp":       testConfig.Networking.ExternalIp,
		"InternalIp":       testConfig.Networking.InternalIp,
		"NetworkName":      testConfig.Networking.ExternalNetwork,
		//"RouteNetworkName": "TestAccVcdVAppVmNet",
		"Catalog":          testSuiteCatalogName,
		"CatalogItem":      testSuiteCatalogOVAItem,
		//"Tags":             "gateway firewall",
	}
}
*/
func TestCreateSgMember(t *testing.T) {
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"ExternalIp":  testConfig.Networking.ExternalIp,
		"InternalIp":  testConfig.Networking.InternalIp,
		"NetworkName": testConfig.Networking.ExternalNetwork,
		//"RouteNetworkName": "TestAccVcdVAppVmNet",
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		//"Tags":             "gateway firewall",
	}

	sgm := createSgMember("Ax1234", "server")
	if sgm.ID != "Ax1234" {
		t.Errorf("Expecting id to be 'Ax1234', but instead got %s", sgm.ID)
	}
	if sgm.Type.TypeName != "server" {
		t.Errorf("Expeted member type to be 'server' but instead got '%s'", sgm.Type.TypeName)
	}
}
