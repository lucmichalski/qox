package pages

import (
	"fmt"

	adminapp "github.com/dfang/qor-demo/app/admin"
	"github.com/dfang/qor-demo/config/db"
	"github.com/dfang/qor-demo/models/blogs"
	"github.com/dfang/qor-demo/utils/funcmapmaker"
	"github.com/qor/admin"
	"github.com/qor/application"
	"github.com/qor/page_builder"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
	"github.com/qor/render"
	"github.com/qor/widget"
)

// New new home app
func New(config *Config) *App {
	return &App{Config: config}
}

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

// ConfigureApplication configure application
func (app App) ConfigureApplication(application *application.Application) {
	controller := &Controller{View: render.New(&render.Config{AssetFileSystem: application.AssetFS.NameSpace("blog")}, "app/pages/views")}
	funcmapmaker.AddFuncMapMaker(controller.View)
	app.ConfigureAdmin(application.Admin)

	application.Router.Get("/blog", controller.Index)
	application.Router.Get("/{article}", controller.Show)
}

// ConfigureAdmin configure admin interface
func (App) ConfigureAdmin(Admin *admin.Admin) {
	// Blog Management
	article := Admin.AddResource(&blogs.Article{}, &admin.Config{Menu: []string{"Pages Management"}})
	article.IndexAttrs("ID", "VersionName", "ScheduledStartAt", "ScheduledEndAt", "Author", "Title")

	// Setup pages
	PageBuilderWidgets := widget.New(&widget.Config{DB: db.DB})
	PageBuilderWidgets.WidgetSettingResource = Admin.NewResource(&adminapp.QorWidgetSetting{}, &admin.Config{Name: "PageBuilderWidgets"})
	PageBuilderWidgets.WidgetSettingResource.NewAttrs(
		&admin.Section{
			Rows: [][]string{{"Kind"}, {"SerializableMeta"}},
		},
	)
	PageBuilderWidgets.WidgetSettingResource.AddProcessor(&resource.Processor{
		Handler: func(value interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
			if widgetSetting, ok := value.(*adminapp.QorWidgetSetting); ok {
				if widgetSetting.Name == "" {
					var count int
					context.GetDB().Set(admin.DisableCompositePrimaryKeyMode, "off").Model(&adminapp.QorWidgetSetting{}).Count(&count)
					widgetSetting.Name = fmt.Sprintf("%v %v", utils.ToString(metaValues.Get("Kind").Value), count)
				}
			}
			return nil
		},
	})
	Admin.AddResource(PageBuilderWidgets, &admin.Config{Menu: []string{"Pages Management"}})

	page := page_builder.New(&page_builder.Config{
		Admin:       Admin,
		PageModel:   &blogs.Page{},
		Containers:  PageBuilderWidgets,
		AdminConfig: &admin.Config{Name: "Pages", Menu: []string{"Pages Management"}, Priority: 1},
	})
	page.IndexAttrs("ID", "Title", "PublishLiveNow")
}
