package azurerm

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func GetLocationField(d *schema.ResourceData) *string {
	return utils.String(azureRMNormalizeLocation(d.Get("location").(string)))
}

func GetTagsField(d *schema.ResourceData) map[string]*string {
	return expandTags(d.Get("tags").(map[string]interface{}))
}

func GetFieldString(d *schema.ResourceData, path string) *string {
	return utils.String(d.Get(path).(string))
}

func SetNameAndRGField(d *schema.ResourceData, name string, rg string) {
	d.Set("name", name)
	d.Set("resource_group_name", rg)
}

func SetLocationField(d *schema.ResourceData, location *string) {
	if location != nil {
		d.Set("location", azureRMNormalizeLocation(*location))
	}
}

func SetTagsField(d *schema.ResourceData, tags map[string]*string) {
	flattenAndSetTags(d, tags)
}

func isZeroValue(v interface{}) bool {
	return v == reflect.Zero(reflect.TypeOf(v)).Interface()
}

func SetFieldOptional(d *schema.ResourceData, field string, v interface{}) {
	if !isZeroValue(v) {
		d.Set(field, reflect.Indirect(reflect.ValueOf(v)).Interface())
	}
}

func SetSubFieldOptional(d map[string]interface{}, field string, v interface{}) {
	if !isZeroValue(v) {
		d[field] = reflect.Indirect(reflect.ValueOf(v)).Interface()
	}
}

func SetFieldObject(d *schema.ResourceData, field string, m interface{}, setChildren func(r map[string]interface{}, v interface{})) {
	output := make(map[string]interface{})
	setChildren(output, m)
	d.Set(field, []interface{}{output})
}

func CreateAPICall(d *schema.ResourceData, c interface{}, ctx context.Context, m string, p interface{}) error {
	name := d.Get("name").(string)
	rg := d.Get("resource_group_name").(string)
	vn := reflect.ValueOf(name)
	vrg := reflect.ValueOf(rg)

	vctx := reflect.ValueOf(ctx)
	vp := reflect.ValueOf(p)

	vc := reflect.ValueOf(c)
	params := []reflect.Value{vctx, vrg, vn, vp}
	returns := vc.MethodByName(m).Call(params)

	verr := returns[1]
	if !verr.IsNil() {
		return verr.Interface().(error)
	}

	params = []reflect.Value{vctx, vrg, vn}
	returns = vc.MethodByName("Get").Call(params)

	verr = returns[1]
	if !verr.IsNil() {
		return verr.Interface().(error)
	}

	vread := returns[0]
	vid := vread.FieldByName("ID")
	if vid.IsNil() {
		return fmt.Errorf("Cannot get the ID %q (Resource Group %q)", name, rg)
	}
	d.SetId(reflect.Indirect(vid).String())

	return nil
}
