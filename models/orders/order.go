package orders

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/dfang/qor-demo/config/db"
	"github.com/dfang/qor-demo/models/users"
	"github.com/gocraft/work"
	"github.com/jinzhu/gorm"
	"github.com/qor/audited"
	"github.com/qor/transition"
)

type Order struct {
	gorm.Model
	audited.AuditedModel

	UserID            *uint
	User              users.User
	PaymentAmount     float32
	PaymentTotal      float32
	AbandonedReason   string
	DiscountValue     uint
	DeliveryMethodID  uint `form:"delivery-method"`
	DeliveryMethod    DeliveryMethod
	PaymentMethod     string
	TrackingNumber    *string
	ShippedAt         *time.Time
	ReturnedAt        *time.Time
	CancelledAt       *time.Time
	ShippingAddressID uint `form:"shippingaddress"`
	ShippingAddress   users.Address
	BillingAddressID  uint `form:"billingaddress"`
	BillingAddress    users.Address
	OrderItems        []OrderItem `json:"order_items"`
	// AmazonAddressAccessToken string
	// AmazonOrderReferenceID   string
	// AmazonAuthorizationID    string
	// AmazonCaptureID          string
	// AmazonRefundID           string
	PaymentLog string `gorm:"size:655250"`

	// 订单来源 0 自家商城, 1 京东
	Source string

	// 订单号 面单号
	OrderNo string `gorm:"unique;not null" json:"order_no"`

	// -- ORDER_TYPE starts with Q 退货的取件单
	OrderType string

	// 客户信息, 从京东后台导入或者扫描枪扫入的
	CustomerAddress string `json:"customer_address"`
	CustomerName    string `json:"customer_name"`
	CustomerPhone   string `json:"customer_phone"`

	// 预约配送时间
	ReservedDeliveryTime string `gorm:"reserved_delivery_time" json:"reserved_delivery_time"`

	// 预约安装时间
	ReservedSetupTime string `gorm:"reserved_setup_time" json:"reserved_setup_time"`

	// 预约取件时间
	ReservedPickupTime string `gorm:"reserved_pickup_time" json:"reserved_pickup_time"`

	// 是否送装一体（这个jd页面抓下来的是什么就存什么, 但是实际上有的订单是非送装一体，如果客户要求，也需要派人安装的，有些订单是取件单)
	// 所以这个字段保持和京东抓下来的一致，另外还要个OrderType， 根据规则或者人工去改OrderType
	IsDeliveryAndSetup string

	// 应收款项
	Receivables float32 `json:"receivables"`

	// 配送员
	ManToDeliverID string     `l10n:"sync"`
	ManToDeliver   users.User `l10n:"sync"`

	// 配送员
	ManToSetupID string
	ManToSetup   users.User `l10n:"sync"`

	// 取件员
	ManToPickupID string     `l10n:"sync"`
	ManToPickup   users.User `l10n:"sync"`

	// 取件费
	PickupFee float32

	// 配送费
	ShippingFee float32

	// 安装费
	SetupFee float32

	transition.Transition
}

func (order Order) ExternalID() string {
	return fmt.Sprint(order.ID)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (order Order) UniqueExternalID() string {
	return fmt.Sprint(order.ID) + "_" + randomString(6)
}

func (order Order) IsCart() bool {
	return order.State == DraftState || order.State == ""
}

func (order Order) Amount() (amount float32) {
	for _, orderItem := range order.OrderItems {
		amount += orderItem.Amount()
	}
	return
}

// DeliveryFee delivery fee
func (order Order) DeliveryFee() (amount float32) {
	return order.DeliveryMethod.Price
}

func (order Order) Total() (total float32) {
	total = order.Amount() - float32(order.DiscountValue)
	total = order.Amount() + float32(order.DeliveryMethod.Price)
	return
}

// AfterCreate 初始状态
func (order *Order) AfterCreate(scope *gorm.Scope) error {
	if strings.Contains(order.OrderNo, "Q") {
		scope.SetColumn("reserved_pickup_time", order.ReservedDeliveryTime)
	}

	var enqueuer = work.NewEnqueuer("qor", db.RedisPool)
	// enqueuer.Enqueue("update_order_items", work.Q{})

	enqueuer.EnqueueIn("create_aftersale", 10, work.Q{"order_no": order.OrderNo})
	return nil
}
