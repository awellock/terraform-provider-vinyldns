/*
Copyright 2018 Comcast Cable Communications Management, LLC
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vinyldns

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vinyldns/go-vinyldns/vinyldns"
)

func TestAccVinylDNSZoneBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccVinylDNSZoneDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccVinylDNSZoneConfigBasic("email@foo.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVinylDNSZoneExists("vinyldns_zone.test_zone"),
					resource.TestCheckResourceAttr("vinyldns_zone.test_zone", "name", "system-test."),
					resource.TestCheckResourceAttr("vinyldns_zone.test_zone", "email", "email@foo.com"),
				),
			},
			resource.TestStep{
				Config: testAccVinylDNSZoneConfigBasic("updated_email@foo.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVinylDNSZoneUpdatedExists("vinyldns_zone.test_zone"),
					resource.TestCheckResourceAttr("vinyldns_zone.test_zone", "name", "system-test."),
					resource.TestCheckResourceAttr("vinyldns_zone.test_zone", "email", "updated_email@foo.com"),
				),
			},
			resource.TestStep{
				ResourceName:      "vinyldns_zone.test_zone",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  testAccVinylDNSZoneImportStateCheck,
			},
		},
	})
}

func testAccVinylDNSZoneImportStateCheck(s []*terraform.InstanceState) error {
	if len(s) != 1 {
		return fmt.Errorf("expected 1 state: %#v", s)
	}

	rs := s[0]

	expName := "system-test."
	name := rs.Attributes["name"]
	if name != expName {
		return fmt.Errorf("expected name attribute to be %s, received %s", expName, name)
	}

	expEmail := "updated_foo@bar.com"
	email := rs.Attributes["email"]
	if email != expEmail {
		return fmt.Errorf("expected email attribute to be %s, received %s", expEmail, email)
	}

	return nil
}

func testAccVinylDNSZoneDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*vinyldns.Client)

	for _, rs := range s.RootModule().Resources {
		log.Printf("[INFO] testing zone destruction; rs.Type: %s", rs.Type)
		if rs.Type != "vinyldns_zone" {
			continue
		}
		id := rs.Primary.ID

		log.Printf("[INFO] testing zone destruction: %s", id)

		// Try to find the zone
		_, err := client.Zone(id)
		if err == nil {
			return fmt.Errorf("Zone still exists")
		}
	}

	return nil
}

func testAccCheckVinylDNSZoneExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckZoneExists("email@foo.com", n, s)
	}
}

func testAccCheckVinylDNSZoneUpdatedExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckZoneExists("updated_email@foo.com", n, s)
	}
}

func testAccCheckZoneExists(email string, n string, s *terraform.State) error {
	rs, ok := s.RootModule().Resources[n]
	if !ok {
		return fmt.Errorf("Not found %s", rs)
	}
	log.Printf("[INFO] testing that zone exists: %s", rs.Primary.ID)

	if rs.Primary.ID == "" {
		return fmt.Errorf("No Zone ID is set")
	}

	client := testAccProvider.Meta().(*vinyldns.Client)

	readZone, err := client.Zone(rs.Primary.ID)
	if err != nil {
		return err
	}

	if readZone.Name != "system-test." {
		return fmt.Errorf("Zone %s not found", "system-test.")
	}

	if readZone.Email != email {
		return fmt.Errorf("Zone %s with email %s not found", "system-test.", email)
	}

	return nil
}

func testAccVinylDNSZoneConfigBasic(email string) string {
	return fmt.Sprintf(`
resource "vinyldns_group" "test_group" {
	name = "terraformtestgroup"
	email = "tftest@tf.com"
	member {
	  id = "ok"
	}
	admin {
	  id = "ok"
	}
}

resource "vinyldns_zone" "test_zone" {
	name = "system-test."
	email = "%s"
	admin_group_id = "${vinyldns_group.test_group.id}"
	depends_on = [
		"vinyldns_group.test_group"
	]
}`, email)
}
