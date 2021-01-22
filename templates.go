package webservice

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
)

type templatesManager struct {
	templates map[string]*template.Template
	lock      sync.RWMutex
	logger    Logger
}

func buildTemplatesManager(
	pagesTemplateDir, pagePattern,
	widgetsTemplateDir, widgetPattern string, log Logger) *templatesManager {

	if log == nil {
		log = &logger{}
	}
	log.Trace("entered...")
	defer log.Trace("done")

	mgr := &templatesManager{
		templates: make(map[string]*template.Template),
		logger:    log,
	}

	mgr.Refresh(pagesTemplateDir, pagePattern,
		widgetsTemplateDir, widgetPattern)

	return mgr
}

func (mgr *templatesManager) RenderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	mgr.logger.Debug("entered...")
	defer mgr.logger.Debug("done")

	mgr.lock.RLock()
	tpl, ok := mgr.templates[name]
	mgr.lock.RUnlock()
	if !ok {
		return fmt.Errorf("the template %v is not existed", name)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tpl.ExecuteTemplate(w, name, data)
}

func (mgr *templatesManager) Refresh(pagesTemplateDir, pagePattern,
	widgetsTemplateDir, widgetPattern string) {

	mgr.logger.Debug("entered...")
	defer mgr.logger.Debug("done")

	templates := make(map[string]*template.Template)
	pagesTemplateDir = strings.TrimSpace(pagesTemplateDir)
	pagePattern = strings.TrimSpace(pagePattern)
	widgetsTemplateDir = strings.TrimSpace(widgetsTemplateDir)
	widgetPattern = strings.TrimSpace(widgetPattern)

	var err error
	pages := []string{}
	widgets := []string{}
	if pagePattern != "" && pagesTemplateDir != "" && ifDirExists(pagesTemplateDir) {
		mgr.logger.Debug(fmt.Sprintf("scan %v page templates in %v", pagePattern, pagesTemplateDir))
		pages, err = filepath.Glob(filepath.Join(pagesTemplateDir, pagePattern))
		if err != nil {
			mgr.logger.Error("filepath.Glob returned error", err)
		} else {
			if widgetPattern != "" && widgetsTemplateDir != "" && ifDirExists(widgetsTemplateDir) {
				mgr.logger.Trace(fmt.Sprintf("scan %v page templates in %v", widgetPattern, widgetsTemplateDir))
				widgets, err = filepath.Glob(filepath.Join(widgetsTemplateDir, widgetPattern))
				if err != nil {
					mgr.logger.Error("filepath.Glob returned error", err)
				}
			}
		}
	}

	for _, page := range pages {
		files := make([]string, 0, len(widgets)+1)
		files = append(files, page)
		files = append(files, widgets...)
		if tpl, err := template.ParseFiles(files...); err == nil {
			templates[filepath.Base(page)] = tpl
			mgr.logger.Debug("parsed html template of", files)
		} else {
			mgr.logger.Trace("parsed html template with error", err)
		}

	}

	mgr.logger.Debug("parsed templates", templates)

	mgr.lock.Lock()
	mgr.templates = templates
	mgr.lock.Unlock()
}
