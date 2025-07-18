package compute_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-google/google/acctest"
	"github.com/hashicorp/terraform-provider-google/google/envvar"
)

func TestAccComputeRoute_defaultInternetGateway(t *testing.T) {
	t.Parallel()

	acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckComputeRouteDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeRoute_defaultInternetGateway(acctest.RandString(t, 10)),
			},
			{
				ResourceName:      "google_compute_route.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccComputeRoute_hopInstance(t *testing.T) {
	instanceName := "tf-test-" + acctest.RandString(t, 10)
	zone := "us-central1-b"

	acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckComputeRouteDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeRoute_hopInstance(instanceName, zone, acctest.RandString(t, 10)),
			},
			{
				ResourceName:      "google_compute_route.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccComputeRoute_resourceManagerTags(t *testing.T) {

	org := envvar.GetTestOrgFromEnv(t)

	routeName := fmt.Sprintf("tf-test-route-resource-manager-tags-%s", acctest.RandString(t, 10))
	tagKeyResult := acctest.BootstrapSharedTestTagKeyDetails(t, "crm-nroute-tagkey", "organizations/"+org, make(map[string]interface{}))
	sharedTagkey, _ := tagKeyResult["shared_tag_key"]
	tagValueResult := acctest.BootstrapSharedTestTagValueDetails(t, "crm-route-tagvalue", sharedTagkey, org)
	context := map[string]interface{}{
		"route_name":   routeName,
		"tag_key_id":   tagKeyResult["name"],
		"tag_value_id": tagValueResult["name"],
	}

	acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckComputeRouteDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeRoute_resourceManagerTags(context),
			},
			{
				ResourceName:            "google_compute_route.acc_route_with_resource_manager_tags",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"params"}, // we don't read tags back. The whole params block is input only
			},
		},
	})
}

func testAccComputeRoute_resourceManagerTags(context map[string]interface{}) string {
	return acctest.Nprintf(`
resource "google_compute_route" "acc_route_with_resource_manager_tags" {
  name             = "%{route_name}"
  dest_range       = "0.0.0.0/0"
  network          = "default"
  next_hop_gateway = "default-internet-gateway"
  priority         = 100
  params {
	resource_manager_tags = {
	  "%{tag_key_id}" = "%{tag_value_id}"
  	}
  }
}
`, context)
}

func testAccComputeRoute_defaultInternetGateway(suffix string) string {
	return fmt.Sprintf(`
resource "google_compute_route" "foobar" {
  name             = "route-test-%s"
  dest_range       = "0.0.0.0/0"
  network          = "default"
  next_hop_gateway = "default-internet-gateway"
  priority         = 100
}
`, suffix)
}

func testAccComputeRoute_hopInstance(instanceName, zone, suffix string) string {
	return fmt.Sprintf(`
data "google_compute_image" "my_image" {
  family  = "debian-11"
  project = "debian-cloud"
}

resource "google_compute_instance" "foo" {
  name         = "%s"
  machine_type = "e2-medium"
  zone         = "%s"

  boot_disk {
    initialize_params {
      image = data.google_compute_image.my_image.self_link
    }
  }

  network_interface {
    network = "default"
  }
}

resource "google_compute_route" "foobar" {
  name                   = "route-test-%s"
  dest_range             = "0.0.0.0/0"
  network                = "default"
  next_hop_instance      = google_compute_instance.foo.name
  next_hop_instance_zone = google_compute_instance.foo.zone
  priority               = 100
}
`, instanceName, zone, suffix)
}
