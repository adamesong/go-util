package struct_tool_test

import (
	"testing"
	"time"

	ptr "github.com/adamesong/go-util/pointer"
	"github.com/adamesong/go-util/struct_tool"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TestStruct struct {
	Name      string
	Age       int
	Id        uuid.UUID
	Price     float64
	Score     decimal.Decimal
	CreatedAt time.Time
	Message   *string
	Amount    *int
	Weight    *float64
	Height    *decimal.Decimal
	ClubId    *uuid.UUID
	UpdatedAt *time.Time
}

// 针对 struct package中的SetUpdateValueWithOption function写测试
// 分别测试CanSetZero为true和为false的情况
// 测试结果：都可以正常修改
func TestSetUpdateValueWithOptionCanSetZeroTrue(t *testing.T) {

	now := time.Now()
	oId := uuid.New()

	newId := uuid.New()

	// 创建一个TestStruct实例
	orginStruct := TestStruct{
		Name:      "John",
		Age:       30,
		Id:        uuid.New(),
		Price:     10.5,
		Score:     decimal.NewFromFloat(85.5),
		CreatedAt: now,
		Message:   ptr.New("orgin message"),
		Amount:    ptr.New(1),
		Weight:    ptr.New(1.1),
		Height:    ptr.New(decimal.NewFromFloat(1.1)),
		ClubId:    &oId,
		UpdatedAt: &now,
	}
	updateStruct := TestStruct{
		Name:      "Jane",
		Age:       35,
		Id:        newId,
		Price:     10.4,
		Score:     decimal.NewFromFloat(85.1),
		CreatedAt: now.Add(time.Hour),
		Message:   ptr.New("update message"),
		Amount:    ptr.New(2),
		Weight:    ptr.New(2.3),
		Height:    ptr.New(decimal.NewFromFloat(2.4)),
		ClubId:    &newId,
		UpdatedAt: ptr.New(now.Add(time.Hour)),
	}
	orginStruct2 := TestStruct{
		Name:      "John",
		Age:       30,
		Id:        uuid.New(),
		Price:     10.5,
		Score:     decimal.NewFromFloat(85.5),
		CreatedAt: now,
		Message:   ptr.New("orgin message"),
		Amount:    ptr.New(1),
		Weight:    ptr.New(1.1),
		Height:    ptr.New(decimal.NewFromFloat(1.1)),
		ClubId:    &oId,
		UpdatedAt: &now,
	}
	zeroStruct := TestStruct{
		Name:      "",
		Age:       0,
		Id:        uuid.Nil,
		Price:     0,
		Score:     decimal.NewFromFloat(0),
		CreatedAt: time.Time{},
		Message:   nil,
		Amount:    nil,
		Weight:    nil,
		Height:    nil,
		ClubId:    nil,
		UpdatedAt: nil,
	}

	// 创建一个Option实例
	option := struct_tool.SetUpdateValueOption{
		CanSetNil: true,
	}

	//! 测试1: 无nil值的情况下，所有值都应被更新
	changed, _, _, err := struct_tool.SetUpdateValueWithOption(&orginStruct, &updateStruct, &option)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if changed == false {
		t.Errorf("Expected changed to be true, but got %v", changed)
	}

	if orginStruct.Age != 35 {
		t.Errorf("Expected Age to be 35, but got %d", orginStruct.Age)
	}
	// 检查是否修改了Name字段的值
	if orginStruct.Name != "Jane" {
		t.Errorf("Expected Name to be 'Jane', but got %s", orginStruct.Name)
	}
	// 检查是否修改了Id字段的值
	if orginStruct.Id != newId {
		t.Errorf("Expected Id to be %v, but got %v", newId, orginStruct.Id)
	}
	if orginStruct.Price != 10.4 {
		t.Errorf("Expected Price to be 10.4, but got %f", orginStruct.Price)
	}
	// 检查是否修改了Score字段的值
	if orginStruct.Score.String() != "85.1" {
		t.Errorf("Expected Score to be 85.1, but got %s", orginStruct.Score.String())
	}
	if orginStruct.CreatedAt != now.Add(time.Hour) {
		t.Errorf("Expected CreatedAt to be %v, but got %v", now.Add(time.Hour), orginStruct.CreatedAt)
	}
	// 检查是否修改了Message字段的值
	if *orginStruct.Message != "update message" {
		t.Errorf("Expected Message to be 'update message', but got %s", *orginStruct.Message)
	}
	// 检查是否修改了Amount字段的值
	if *orginStruct.Amount != 2 {
		t.Errorf("Expected Amount to be 2, but got %d", *orginStruct.Amount)
	}
	if *orginStruct.Weight != 2.3 {
		t.Errorf("Expected Weight to be 2.3, but got %f", *orginStruct.Weight)
	}
	// 检查是否修改了Height字段的值
	if orginStruct.Height.String() != "2.4" {
		t.Errorf("Expected Height to be 2.4, but got %s", orginStruct.Height.String())
	}
	if orginStruct.ClubId.String() != newId.String() {
		t.Errorf("Expected ClubId to be %s, but got %s", newId.String(), orginStruct.ClubId.String())
	}
	// 检查是否修改了UpdatedAt字段的值
	if *orginStruct.UpdatedAt != now.Add(time.Hour) {
		t.Errorf("Expected UpdatedAt to be %v, but got %v", now.Add(time.Hour), *orginStruct.UpdatedAt)
	}

	//! 测试2: 除nil值以外的零值都被更新，nil值由于设置了CanSetNil所以也将被更新

	changed, _, _, err = struct_tool.SetUpdateValueWithOption(&orginStruct, &zeroStruct, &option)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !changed {
		t.Errorf("Expected changed to be true, but got %v", changed)
	}

	if orginStruct.Age != 0 {
		t.Errorf("Expected Age to be 0, but got %d", orginStruct.Age)
	}
	// 检查是否修改了Name字段的值
	if orginStruct.Name != "" {
		t.Errorf("Expected Name to be '', but got %s", orginStruct.Name)
	}

	if orginStruct.Id != uuid.Nil {
		t.Errorf("Expected Id to be %v, but got %v", uuid.UUID{}, orginStruct.Id)
	}
	// 检查是否修改了Price字段的值
	if orginStruct.Price != 0 {
		t.Errorf("Expected Price to be 0, but got %f", orginStruct.Price)
	}
	if orginStruct.Score.String() != "0" {
		t.Errorf("Expected Score to be 0, but got %s", orginStruct.Score.String())
	}
	// 检查是否修改了CreatedAt字段的值
	if !orginStruct.CreatedAt.IsZero() {
		t.Errorf("Expected CreatedAt to be %v, but got %v", time.Time{}, orginStruct.CreatedAt)
	}
	if orginStruct.Message != nil {
		t.Errorf("Expected Message to be nil, but got %s", *orginStruct.Message)
	}
	// 检查是否修改了Amount字段的值
	if orginStruct.Amount != nil {
		t.Errorf("Expected Amount to be nil, but got %d", *orginStruct.Amount)
	}
	if orginStruct.Weight != nil {
		t.Errorf("Expected Weight to be nil, but got %f", *orginStruct.Weight)
	}
	// 检查是否修改了Height字段的值
	if orginStruct.Height != nil {
		t.Errorf("Expected Height to be nil, but got %s", orginStruct.Height.String())
	}
	if orginStruct.ClubId != nil {
		t.Errorf("Expected ClubId to be nil, but got %s", orginStruct.ClubId.String())
	}
	// 检查是否修改了UpdatedAt字段的值
	if orginStruct.UpdatedAt != nil {
		t.Errorf("Expected UpdatedAt to be nil, but got %v", *orginStruct.UpdatedAt)
	}

	//! 测试3: 除nil值以外的零值都被更新，nil值由于设置了CanSetNil = false 所以也将不会被更新

	option2 := struct_tool.SetUpdateValueOption{
		CanSetNil: false,
	}

	// 去除了price和height，这两个字段相当于默认也是0值，也将被更新；其他在partialZeroStruct中的nil值将不会被更新
	partialZeroStruct := TestStruct{
		Name: "Jane",
		Age:  0,
		Id:   newId,

		Score:     decimal.NewFromFloat(0),
		CreatedAt: time.Time{},
		Message:   nil,
		Amount:    nil,
		Weight:    nil,

		ClubId:    nil,
		UpdatedAt: nil,
	}
	changed, _, _, err = struct_tool.SetUpdateValueWithOption(&orginStruct2, &partialZeroStruct, &option2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if changed == false {
		t.Errorf("Expected changed to be true, but got %v", changed)
	}
	// 检查是否修改了Name字段的值
	if orginStruct2.Name != "Jane" {
		t.Errorf("Expected Name to be 'Jane', but got %s", orginStruct2.Name)
	}
	if orginStruct2.Id != newId {
		t.Errorf("Expected Id to be %v, but got %v", newId, orginStruct2.Id)
	}
	// 检查是否修改了Price字段的值
	if orginStruct2.Price != 0 {
		t.Errorf("Expected Price to be 0, but got %f", orginStruct2.Price)
	}
	if orginStruct2.Score.String() != "0" {
		t.Errorf("Expected Score to be 0, but got %s", orginStruct2.Score.String())
	}
	// 检查是否修改了CreatedAt字段的值
	if !orginStruct2.CreatedAt.IsZero() {
		t.Errorf("Expected CreatedAt to be %v, but got %v", time.Time{}, orginStruct2.CreatedAt)
	}
	if orginStruct2.Message == nil {
		t.Errorf("Expected Message to be nil, but got %s", *orginStruct2.Message)
	}

	// 检查是否修改了Amount字段的值
	if orginStruct2.Amount == nil {
		t.Errorf("Expected Amount to be nil, but got %d", *orginStruct2.Amount)
	}
	if orginStruct2.Weight == nil {
		t.Errorf("Expected Weight to be nil, but got %f", *orginStruct2.Weight)
	}

	// 检查是否修改了Height字段的值
	if orginStruct2.Height == nil {
		t.Errorf("Expected Height to be not nil, but got nil")
	}
	// 检查是否修改了ClubId字段的值
	if orginStruct2.ClubId == nil {
		t.Errorf("Expected ClubId to be nil, but got %s", orginStruct2.ClubId.String())
	}
	if orginStruct2.UpdatedAt == nil {
		t.Errorf("Expected UpdatedAt to be nil, but got %v", *orginStruct2.UpdatedAt)
	}

	//! 测试4: 两个不同的结构体，同名field的类型不同，前者是值，后者是指针
	orgin4 := struct {
		Name string
	}{Name: "john"}
	update4 := struct {
		Name *string
	}{Name: ptr.New("update name")}
	option4 := struct_tool.SetUpdateValueOption{
		CanSetNil: true,
	}
	changed, _, _, err = struct_tool.SetUpdateValueWithOption(&orgin4, &update4, &option4)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if changed == false {
		t.Errorf("Expected changed to be true, but got %v", changed)
	}
	// 检查是否修改了Name字段的值
	if orgin4.Name != "update name" {
		t.Errorf("Expected Name to be 'update name', but got %s", orgin4.Name)
	}

	//! 测试5: 两个不同的结构体，同名field的类型不同，前者是指针，后者是值
	orgin5 := struct {
		Name *string
	}{Name: ptr.New("john")}
	update5 := struct {
		Name string
	}{Name: "update name"}
	option5 := struct_tool.SetUpdateValueOption{
		CanSetNil: true,
	}
	changed, _, _, err = struct_tool.SetUpdateValueWithOption(&orgin5, &update5, &option5)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if changed == false {
		t.Errorf("Expected changed to be true, but got %v", changed)
	}
	if *orgin5.Name != "update name" {
		t.Errorf("Expected Name to be 'update name', but got %s", *orgin5.Name)
	}
	//! 测试6: 两个不同的结构体，同名field的类型不同，前者是指针，后者是指针
	orgin6 := struct {
		Name *string
	}{Name: ptr.New("john")}
	update6 := struct {
		Name *string
	}{Name: ptr.New("update name")}
	option6 := struct_tool.SetUpdateValueOption{
		CanSetNil: true,
	}
	changed, _, _, err = struct_tool.SetUpdateValueWithOption(&orgin6, &update6, &option6)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if changed == false {
		t.Errorf("Expected changed to be true, but got %v", changed)
	}
	if *orgin6.Name != "update name" {
		t.Errorf("Expected Name to be 'update name', but got %s", *orgin6.Name)
	}
	//! 测试7: 两个不同的结构体，同名field的类型不同，前者是值且为非零，后者是指针且为nil
	orgin7 := struct {
		Name string
	}{Name: "john"}
	zero7 := struct {
		Name *string
	}{Name: nil}
	option7 := struct_tool.SetUpdateValueOption{
		CanSetNil: true,
	}
	changed, _, _, err = struct_tool.SetUpdateValueWithOption(&orgin7, &zero7, &option7)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if changed == true {
		t.Errorf("Expected changed to be true, but got %v", changed)
	}

	//! 测试8: 两个不同的结构体，同名field的类型不同，前者是指针且为nil，后者是值且为零
	orgin8 := struct {
		Name *string
	}{Name: nil}
	zero8 := struct {
		Name string
	}{Name: ""}
	option8 := struct_tool.SetUpdateValueOption{
		CanSetNil: true,
	}
	changed, _, _, err = struct_tool.SetUpdateValueWithOption(&orgin8, &zero8, &option8)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if changed == false {
		t.Errorf("Expected changed to be true, but got %v", changed)
	}

	if *orgin8.Name != "" {
		t.Errorf("Expected Name to be '', but got %s", *orgin8.Name)
	}

}
