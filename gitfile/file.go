package gitfile

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/hashicorp/terraform/helper/schema"
)

func fileResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"checkout_dir": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"path": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"contents": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
		Create: fileCreateUpdate,
		Read:   fileRead,
		Delete: fileDelete,
	}
}

func fileCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	checkout_dir := d.Get("checkout_dir").(string)
	log.Printf("[DEBUG] Creating/updateing gitfile file at %q", checkout_dir)
	lockCheckout(checkout_dir)
	defer unlockCheckout(checkout_dir)

	filePath := path.Join(checkout_dir, d.Get("path").(string))
	contents := d.Get("contents").(string)

	if err := os.MkdirAll(path.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("Failed to create parent directory: %s", err)
	}
	if err := ioutil.WriteFile(filePath, []byte(contents), 0666); err != nil {
		return fmt.Errorf("Failed to write file to %s: %s", filePath, err)
	}

	d.SetId(filePath)

	return nil
}

func fileRead(d *schema.ResourceData, meta interface{}) error {
	checkout_dir := d.Get("checkout_dir").(string)
	log.Printf("[DEBUG] Reading gitfile file from %q", checkout_dir)
	lockCheckout(checkout_dir)
	defer unlockCheckout(checkout_dir)

	if _, err := os.Stat(checkout_dir); err != nil {
		d.SetId("")
		return nil
	}

	filePath := d.Get("path").(string)
	contents, err := gitCommand(checkout_dir, "show", "HEAD:"+filePath)
	if err != nil {
		return fmt.Errorf("Failed to execute git show of %q: %s", filePath, err)
	}

	log.Printf("[DEBUG] Setting file contents to %q", contents)

	err = d.Set("contents", string(contents))
	if err != nil {
		return err
	}

	return nil
}

func fileDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
