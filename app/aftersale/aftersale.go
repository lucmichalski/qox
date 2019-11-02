package aftersale

// "net/http"
import (
	"fmt"
	"time"

	"github.com/dfang/qor-demo/config/db"
	"github.com/dfang/qor-demo/models/aftersales"
	"github.com/dfang/qor-demo/models/settings"
	"github.com/dfang/qor-demo/models/users"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/now"
	"github.com/qor/activity"
	"github.com/qor/admin"
	"github.com/qor/application"
	"github.com/qor/qor"
)

// New new home app
func New(config *Config) *App {
	return &App{Config: config}
}

// NewWithDefault new home app
func NewWithDefault() *App {
	return &App{Config: &Config{}}
}

// App home app
type App struct {
	Config *Config
}

// Config home config struct
type Config struct {
}

var brands []settings.Brand
var service_types []settings.ServiceType
var workmen []users.User

// ConfigureApplication configure application
func (app App) ConfigureApplication(application *application.Application) {
	// 售后后台
	app.ConfigureAdmin(application.Admin)
}

// ConfigureAdmin configure admin interface
func (App) ConfigureAdmin(Admin *admin.Admin) {

	db.DB.Select("name, id").Find(&brands)
	db.DB.Select("name, id").Find(&service_types)
	db.DB.Select("name, id").Where("role = ?", "workman").Find(&workmen)

	aftersale := Admin.AddResource(&aftersales.Aftersale{}, &admin.Config{Menu: []string{"Aftersale Management"}, Priority: 1})
	aftersale.Meta(&admin.Meta{Name: "User", Type: "aftersale_user_field"})

	manufacturer := Admin.AddResource(&aftersales.Manufacturer{}, &admin.Config{Menu: []string{"Aftersale Management"}, Priority: 4})
	Admin.AddResource(&settings.Brand{}, &admin.Config{Name: "Brand", Menu: []string{"Aftersale Management"}, Priority: 3})
	Admin.AddResource(&settings.ServiceType{}, &admin.Config{Name: "ServiceType", Menu: []string{"Aftersale Management"}, Priority: 2})
	Admin.AddResource(&users.WechatProfile{}, &admin.Config{Name: "WechatProfile", Menu: []string{"Aftersale Management"}, Priority: 5})

	activity.Register(aftersale)

	settlement := Admin.AddResource(&aftersales.Settlement{}, &admin.Config{Menu: []string{"Settlement Management"}, Priority: 2})
	settlement.IndexAttrs("-CreatedBy", "-UpdatedBy", "ID", "Workman", "Amount", "Direction", "Aftersale", "State", "CreatedAt")
	settlement.Meta(&admin.Meta{
		Name:       "Direction",
		Type:       "select_one",
		Collection: []string{"收入", "提现", "罚款", "奖励"},
	})
	// 虚拟field， 仅为在列表页正确显示师傅姓名和链接，又不影响form里的下拉框
	settlement.Meta(&admin.Meta{
		Name: "Workman",
		Type: "settlement_user_field",
		Valuer: func(record interface{}, context *qor.Context) interface{} {
			if p, ok := record.(*aftersales.Settlement); ok {
				// fmt.Println("ok")
				var user users.User
				context.GetDB().Model(users.User{}).Where("id = ?", p.UserID).Find(&user)
				// fmt.Println(p)
				// fmt.Println(p.User)
				return user
			}

			return record
		},
	})

	settlement.Meta(&admin.Meta{
		Name: "User",
		Type: "select_one",
		Collection: func(_ interface{}, context *admin.Context) (options [][]string) {
			var users []users.User
			context.GetDB().Where("role = ?", "workman").Find(&users)

			for _, n := range users {
				idStr := fmt.Sprintf("%d", n.ID)
				var option = []string{idStr, n.Name}
				options = append(options, option)
			}

			return options
		},
	})
	settlement.Meta(&admin.Meta{Name: "State", Type: "string", FormattedValuer: func(record interface{}, _ *qor.Context) (result interface{}) {
		m := record.(*aftersales.Settlement)
		switch m.State {
		case "frozen":
			return "冻结中"
		case "free":
			return "已解冻"
		case "withdrawed":
			return "已提现"
		default:
			// return "N/A"
			return m.State
		}
	}})
	// settlement.Meta(&admin.Meta{Name: "Amount", Type: "float32", FormattedValuer: func(record interface{}, _ *qor.Context) (result interface{}) {
	// 	m := record.(*aftersales.Settlement)
	// 	if m.Amount < 0 {
	// 		return (-m.Amount)
	// 	}
	// 	return m.Amount
	// }})

	balance := Admin.AddResource(&aftersales.Balance{}, &admin.Config{Menu: []string{"Settlement Management"}, Priority: 1})
	balance.IndexAttrs("-ID", "-CreatedAt", "-CreatedBy", "-UpdatedBy", "User", "FrozenAmount", "FreeAmount", "TotalAmount", "WithdrawAmount", "UpdatedAt")
	balance.Meta(&admin.Meta{Name: "WithdrawAmount", Type: "float32", FormattedValuer: func(record interface{}, _ *qor.Context) (result interface{}) {
		m := record.(*aftersales.Balance)
		if m.WithdrawAmount < 0 {
			return (-m.WithdrawAmount)
		}
		return m.WithdrawAmount
	}})
	balance.Meta(&admin.Meta{Name: "User", Type: "balance_user_field", Label: "师傅"})

	aftersale.Meta(&admin.Meta{
		Name: "ServiceType",
		Type: "select_one",
		// Collection: []string{"安装", "维修", "清洗"},
		Collection: func(value interface{}, context *qor.Context) (options [][]string) {
			for _, m := range service_types {
				idStr := fmt.Sprintf("%s", m.Name)
				var option = []string{idStr, m.Name}
				options = append(options, option)
			}
			return options
		},
	})

	aftersale.Meta(&admin.Meta{
		Name: "Source",
		Type: "select_one",
		Collection: func(value interface{}, context *qor.Context) (options [][]string) {
			for _, m := range brands {
				idStr := fmt.Sprintf("%s", m.Name)
				var option = []string{idStr, m.Name}
				options = append(options, option)
			}
			return options
		},
	})

	// https://doc.getqor.com/admin/metas/select-one.html
	// Generate options by data from the database
	aftersale.Meta(&admin.Meta{
		Name:  "UserID",
		Type:  "select_one",
		Label: "分配",
		Config: &admin.SelectOneConfig{
			Collection: func(_ interface{}, context *admin.Context) (options [][]string) {
				var users []users.User
				context.GetDB().Where("role = ?", "workman").Find(&users)
				for _, n := range users {
					idStr := fmt.Sprintf("%d", n.ID)
					var option = []string{idStr, n.Name}
					options = append(options, option)
				}
				return options
			},
			AllowBlank: true,
		}})

	manufacturer.Action(&admin.Action{
		Name: "打开厂家后台网站",
		URL: func(record interface{}, context *admin.Context) string {
			if item, ok := record.(*aftersales.Manufacturer); ok {
				return fmt.Sprintf("%s", item.URL)
			}
			return "#"
		},
		URLOpenType: "_blank",
		Modes:       []string{"menu_item", "edit", "show"},
	})

	configureMetas(aftersale)
	configureActions(Admin, aftersale)
	configureScopes(aftersale)

	configureScopesForSettlements(settlement)
	configureScopesForBalances(balance)

	// aftersale.UseTheme("grid")
	// aftersale.UseTheme("publish2")
	aftersale.UseTheme("fancy")

	// aftersale.FindManyHandler = func(results interface{}, context *qor.Context) error {
	// 	db             = context.GetDB()
	// 	scope          = db.NewScope(record)

	// 	// find records and decode them to results
	// 	return nil
	// }
}

func configureMetas(model *admin.Resource) {
	model.EditAttrs("-UserID", "-User", "-CreatedAt", "-UpdatedAt", "-CreatedBy", "-UpdatedBy", "-State")
	model.NewAttrs("-UserID", "-User", "-CreatedAt", "-UpdatedAt", "-CreatedBy", "-UpdatedBy", "-State")
	model.IndexAttrs("-UserID", "-CreatedAt", "-UpdatedAt", "-CreatedBy", "-UpdatedBy", "-Remark", "-ServiceContent", "-ReservedServiceTime", "-Source")
	model.Meta(&admin.Meta{Name: "State", Type: "string", FormattedValuer: func(record interface{}, _ *qor.Context) (result interface{}) {
		m := record.(*aftersales.Aftersale)
		switch m.State {
		case "created":
			return "已接收"
		case "inquired":
			return "已预约"
		case "scheduled":
			return "已指派"
		case "overdue":
			return "已超时"
		case "audited":
			return "已审核"
		case "processing":
			return "待上门"
		case "processed":
			return "已服务"
		default:
			// return "N/A"
			return m.State
		}
	}})
}

func configureScopes(model *admin.Resource) {
	// filter by order state
	for _, item := range aftersales.STATES {
		var item = item // 这句必须有否则会报错，永远都是最后一个值
		model.Scope(&admin.Scope{
			Name:  item,
			Label: item,
			Group: "Filter By State",
			Handler: func(db *gorm.DB, context *qor.Context) *gorm.DB {
				// 两种写法都可以
				// return db.Where(aftersales.Aftersale{State: item})
				// if item == "overdue" {
				// return db.Where("state = 'scheduled'").Where("updated_at <= NOW() - INTERVAL '20 minutes'")
				// return db.Where("state = 'scheduled'")
				// }
				return db.Where("state = ?", item)
			},
		})
	}

	// filter by order type
	for _, item := range service_types {
		var item = item // 这句必须有否则会报错，永远都是最后一个值
		model.Scope(&admin.Scope{
			Name:  item.Name,
			Label: item.Name,
			Group: "Filter By Type",
			Handler: func(db *gorm.DB, context *qor.Context) *gorm.DB {
				// 两种写法都可以
				return db.Where("service_type = ?", item.Name)
			},
		})
	}

	// filter by order source
	// var brands = []settings.Brand{
	// 	settings.Brand{
	// 		Name: "海尔",
	// 	},
	// 	settings.Brand{
	// 		Name: "格力",
	// 	},
	// }
	for _, item := range brands {
		var item = item
		model.Scope(&admin.Scope{
			Name:  item.Name,
			Label: item.Name,
			Group: "Filter By Source",
			Handler: func(db *gorm.DB, context *qor.Context) *gorm.DB {
				return db.Where("source = ?", item.Name)
			},
		})
	}

	for _, item := range workmen {
		var item = item
		model.Scope(&admin.Scope{
			Name:  item.Name,
			Label: item.Name,
			Group: "Filter By Workman",
			Handler: func(db *gorm.DB, context *qor.Context) *gorm.DB {
				// 两种写法都可以
				return db.Where("user_id = ?", item.ID)
			},
		})
	}

	model.Filter(&admin.Filter{
		Name: "created_at",
		Config: &admin.DatetimeConfig{
			ShowTime: false,
		},
	})

	// define scopes for Order
	model.Scope(&admin.Scope{
		Name:  "Today",
		Label: "Today",
		// Default: true, // 如果开启了这个，那么所有的Actions With User Input 就会废掉, 因为argument.FindSelectedRecords 返回为空
		Group: "Filter By Date",
		Handler: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("created_at >= ?", now.BeginningOfDay()).Where("created_at <=? ", time.Now())
		},
	})
	model.Scope(&admin.Scope{
		Name:  "Yesterday",
		Label: "Yesterday",
		Group: "Filter By Date",
		Handler: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			now.WeekStartDay = time.Monday
			// select order_no, customer_name, item_name::varchar(20), quantity, created_at
			// from orders_view
			// where created_at between now() - interval '2 day' and  now() - interval '1 day';
			// return db.Where("created_at between now() - interval '2 day' and  now() - interval '1 day'")
			return db.Where("created_at >= ?", now.BeginningOfDay().AddDate(0, 0, -1)).Where("created_at <=? ", now.EndOfDay().AddDate(0, 0, -1))
		},
	})
	model.Scope(&admin.Scope{
		Name:  "ThisWeek",
		Label: "This Week",
		Group: "Filter By Date",
		Handler: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			now.WeekStartDay = time.Monday
			return db.Where("created_at >= ?", now.BeginningOfWeek()).Where("created_at <=? ", now.EndOfWeek())
		},
	})
	model.Scope(&admin.Scope{
		Name:  "ThisMonth",
		Label: "This Month",
		Group: "Filter By Date",
		Handler: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			now.WeekStartDay = time.Monday
			return db.Where("created_at >= ?", now.BeginningOfMonth()).Where("created_at <=? ", now.EndOfMonth())
		},
	})
	model.Scope(&admin.Scope{
		Name:  "ThisQuarter",
		Label: "This Quarter",
		Group: "Filter By Date",
		Handler: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("created_at >= ?", now.BeginningOfQuarter()).Where("created_at <=? ", now.EndOfQuarter())
		},
	})
	model.Scope(&admin.Scope{
		Name:  "ThisYear",
		Label: "This Year",
		Group: "Filter By Date",
		Handler: func(db *gorm.DB, context *qor.Context) *gorm.DB {
			return db.Where("created_at >= ?", now.BeginningOfYear()).Where("created_at <=? ", now.EndOfYear())
		},
	})
}

func configureActions(Admin *admin.Admin, aftersale *admin.Resource) {
	// 预约客户
	type reserveActionArgument struct {
		Remark string
	}
	reserveActionArgumentResource := Admin.NewResource(&reserveActionArgument{})
	aftersale.Action(&admin.Action{
		Name: "预约",
		Handler: func(argument *admin.ActionArgument) error {
			var (
				arg = argument.Argument.(*reserveActionArgument)
				tx  = argument.Context.GetDB()
			)
			for _, record := range argument.FindSelectedRecords() {
				item := record.(*aftersales.Aftersale)
				aftersales.OrderStateMachine.Trigger("inquire", item, tx, "from created to inquired")
				item.Remark = arg.Remark
				tx.Model(aftersales.Aftersale{}).Save(&item)
			}
			return nil
		},
		Visible: func(record interface{}, context *admin.Context) bool {
			if item, ok := record.(*aftersales.Aftersale); ok {
				return item.State == "created"
			}
			return false
			// return true
		},
		Resource: reserveActionArgumentResource,
		Modes:    []string{"edit", "show", "menu_item"},
	})

	// 指派师傅
	type setupActionArgument struct {
		UserID uint
	}
	setupActionArgumentResource := Admin.NewResource(&setupActionArgument{})
	setupActionArgumentResource.Meta(&admin.Meta{
		Name: "UserID",
		Type: "select_one",
		// Valuer: func(record interface{}, context *qor.Context) interface{} {
		// 	// return record.(*users.User).ID
		// 	return ""
		// },
		Collection: func(value interface{}, context *qor.Context) (options [][]string) {
			var setupMen []users.User
			context.GetDB().Where("role = ?", "workman").Find(&setupMen)
			// context.GetDB().Find(&setupMen)
			for _, m := range setupMen {
				idStr := fmt.Sprintf("%d", m.ID)
				var option = []string{idStr, m.Name}
				options = append(options, option)
			}
			return options
		},
		// Collection: []string{"Male", "Female", "Unknown"},
	})
	aftersale.Action(&admin.Action{
		Name: "指派",
		Handler: func(argument *admin.ActionArgument) error {
			var (
				tx  = argument.Context.GetDB().Begin()
				arg = argument.Argument.(*setupActionArgument)
			)
			for _, record := range argument.FindSelectedRecords() {
				// argument.Context.GetDB().Model(record).UpdateColumn("user_id", arg.UserID)
				// argument.Context.GetDB().Model(record).UpdateColumn("state", "scheduled")
				item := record.(*aftersales.Aftersale)
				item.UserID = arg.UserID
				aftersales.OrderStateMachine.Trigger("schedule", item, tx, "scheduled to user_id: "+fmt.Sprintf("%d", arg.UserID))
				// 无论是inquired、scheduled，还是overdue状态，通过指派按钮重新指派的时候都要变为schedued状态
				// item.State = "scheduled" // 直接这样不触发event
				if err := tx.Save(item).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
			tx.Commit()
			return nil
		},
		Visible: func(record interface{}, context *admin.Context) bool {
			if item, ok := record.(*aftersales.Aftersale); ok {
				return item.State == "inquired"
			}
			return true
		},
		Resource: setupActionArgumentResource,
		Modes:    []string{"edit", "show", "menu_item"},
	})

	aftersale.Action(&admin.Action{
		Name: "重新指派",
		Handler: func(argument *admin.ActionArgument) error {
			var (
				tx  = argument.Context.GetDB().Begin()
				arg = argument.Argument.(*setupActionArgument)
			)
			for _, record := range argument.FindSelectedRecords() {
				// argument.Context.GetDB().Model(record).UpdateColumn("user_id", arg.UserID)
				item := record.(*aftersales.Aftersale)
				item.UserID = arg.UserID
				aftersales.OrderStateMachine.Trigger("reschedule", item, tx, "rescheduled to user_id: "+fmt.Sprintf("%d", arg.UserID))
				// aftersales.OrderState.Trigger("schedule", item, tx, "scheduled to user_id: "+fmt.Sprintf("%d", arg.UserID))
				// 无论是inquired、scheduled，还是overdue状态，通过指派按钮重新指派的时候都要变为schedued状态
				// item.State = "scheduled"
				if err := tx.Save(item).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
			tx.Commit()
			return nil
		},
		Visible: func(record interface{}, context *admin.Context) bool {
			if item, ok := record.(*aftersales.Aftersale); ok {
				return item.State == "scheduled" || item.State == "overdue"
			}
			return true
		},
		Resource: setupActionArgumentResource,
		Modes:    []string{"edit", "show", "menu_item"},
	})

	// // 提示用户
	// type notifyCustomerArgument struct {
	// 	Content string
	// }
	// notifyCustomerArgumentResource := Admin.NewResource(&notifyCustomerArgument{})
	// aftersale.Action(&admin.Action{
	// 	Name: "提示用户",
	// 	Handler: func(argument *admin.ActionArgument) error {
	// 		var (
	// 		// arg = argument.Argument.(*setupActionArgument)
	// 		)
	// 		// for _, record := range argument.FindSelectedRecords() {
	// 		// 	// 给用户发短信
	// 		// }
	// 		return nil
	// 	},
	// 	Visible: func(record interface{}, context *admin.Context) bool {
	// 		// if item, ok := record.(*aftersales.Aftersale); ok {
	// 		// 	return item.State == "inquired"
	// 		// }
	// 		return true
	// 	},
	// 	Resource: notifyCustomerArgumentResource,
	// 	Modes:    []string{"edit", "show", "menu_item"},
	// })

	// // 提示师傅
	// type notifyWorkerArgument struct {
	// 	Content string
	// }
	// notifyWorkerArgumentResource := Admin.NewResource(&setupActionArgument{})
	// aftersale.Action(&admin.Action{
	// 	Name: "提示师傅",
	// 	Handler: func(argument *admin.ActionArgument) error {
	// 		var (
	// 		// arg = argument.Argument.(*setupActionArgument)
	// 		)
	// 		// for _, record := range argument.FindSelectedRecords() {
	// 		// 	// 给用户发短信
	// 		// }
	// 		return nil
	// 	},
	// 	Visible: func(record interface{}, context *admin.Context) bool {
	// 		// if item, ok := record.(*aftersales.Aftersale); ok {
	// 		// 	return item.State == "inquired"
	// 		// }
	// 		return true
	// 	},
	// 	Resource: notifyWorkerArgumentResource,
	// 	Modes:    []string{"edit", "show", "menu_item"},
	// })

	// 审核
	type auditActionArgument struct {
		Fee float32
	}
	auditActionArgumentResource := Admin.NewResource(&auditActionArgument{})
	aftersale.Action(&admin.Action{
		Name: "审核通过",
		Handler: func(argument *admin.ActionArgument) error {
			var (
				tx  = argument.Context.GetDB().Begin()
				arg = argument.Argument.(*auditActionArgument)
			)
			for _, record := range argument.FindSelectedRecords() {
				item := record.(*aftersales.Aftersale)
				item.Fee = arg.Fee
				aftersales.OrderStateMachine.Trigger("audit", item, tx, "audit ok")
				if err := tx.Save(item).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
			tx.Commit()
			return nil
		},
		Visible: func(record interface{}, context *admin.Context) bool {
			if item, ok := record.(*aftersales.Aftersale); ok {
				return item.State == "processed"
			}
			return true
		},
		Resource: auditActionArgumentResource,
		Modes:    []string{"edit", "show", "menu_item"},
	})
	aftersale.Action(&admin.Action{
		Name: "审核不通过",
		Handler: func(argument *admin.ActionArgument) error {
			var (
				tx = argument.Context.GetDB().Begin()
			)
			for _, record := range argument.FindSelectedRecords() {
				item := record.(*aftersales.Aftersale)
				aftersales.OrderStateMachine.Trigger("audit_failed", item, tx, "audit failed")
				if err := tx.Save(item).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
			tx.Commit()
			return nil
		},
		Visible: func(record interface{}, context *admin.Context) bool {
			if item, ok := record.(*aftersales.Aftersale); ok {
				return item.State == "processed"
			}
			return true
		},
		Modes: []string{"edit", "show", "menu_item"},
	})
}

func configureScopesForSettlements(model *admin.Resource) {
	for _, item := range workmen {
		var item = item
		model.Scope(&admin.Scope{
			Name:  item.Name,
			Label: item.Name,
			Group: "Filter By Workman",
			Handler: func(db *gorm.DB, context *qor.Context) *gorm.DB {
				// 两种写法都可以
				return db.Where("user_id = ?", item.ID)
			},
		})
	}
}

func configureScopesForBalances(model *admin.Resource) {
	for _, item := range workmen {
		var item = item
		model.Scope(&admin.Scope{
			Name:  item.Name,
			Label: item.Name,
			Group: "Filter By Workman",
			Handler: func(db *gorm.DB, context *qor.Context) *gorm.DB {
				// 两种写法都可以
				return db.Where("user_id = ?", item.ID)
			},
		})
	}
}
