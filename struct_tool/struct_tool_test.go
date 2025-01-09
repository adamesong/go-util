package struct_tool_test

import (
	"fmt"
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
		CanSetZero: true,
	}

	// 调用SetUpdateValueWithOption函数
	changed, originMap, changeMap := struct_tool.SetUpdateValueWithOption(&orginStruct, &updateStruct, &option)
	if changed == false {
		t.Errorf("Expected changed to be true, but got %v", changed)
	}
	fmt.Println("originMap:", originMap)
	fmt.Println("changeMap:", changeMap)
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

	// 测试更新为零值
	changed, originMap, changeMap = struct_tool.SetUpdateValueWithOption(&orginStruct, &zeroStruct, &option)
	if changed == false {
		t.Errorf("Expected changed to be true, but got %v", changed)
	}
	fmt.Println("originMap:", originMap)
	fmt.Println("changeMap:", changeMap)

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

	option2 := struct_tool.SetUpdateValueOption{
		CanSetZero: false,
	}
	changed, _, _ = struct_tool.SetUpdateValueWithOption(&orginStruct, &zeroStruct, &option2)
	if changed == true {
		t.Errorf("Expected changed to be false, but got %v", changed)
	}

}
