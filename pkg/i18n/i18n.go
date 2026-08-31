package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// I18n 国际化管理器
type I18n struct {
	defaultLang string
	translations map[string]map[string]string
}

// New 创建国际化管理器
func New(defaultLang string) *I18n {
	return &I18n{
		defaultLang:  defaultLang,
		translations: make(map[string]map[string]string),
	}
}

// LoadTranslations 加载翻译文件
func (i *I18n) LoadTranslations(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}

		lang := strings.TrimSuffix(name, ".json")
		path := filepath.Join(dir, name)

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var translations map[string]string
		if err := json.Unmarshal(data, &translations); err != nil {
			return err
		}

		i.translations[lang] = translations
	}

	return nil
}

// T 翻译文本
func (i *I18n) T(lang, key string, args ...interface{}) string {
	translations, ok := i.translations[lang]
	if !ok {
		translations = i.translations[i.defaultLang]
	}

	if translations == nil {
		return key
	}

	value, ok := translations[key]
	if !ok {
		// 尝试默认语言
		if lang != i.defaultLang {
			defaultTranslations := i.translations[i.defaultLang]
			if defaultTranslations != nil {
				if v, ok := defaultTranslations[key]; ok {
					return v
				}
			}
		}
		return key
	}

	return value
}

// GetTranslations 获取指定语言的所有翻译
func (i *I18n) GetTranslations(lang string) map[string]string {
	translations, ok := i.translations[lang]
	if !ok {
		return i.translations[i.defaultLang]
	}
	return translations
}

// GetSupportedLanguages 获取支持的语言列表
func (i *I18n) GetSupportedLanguages() []string {
	languages := make([]string, 0, len(i.translations))
	for lang := range i.translations {
		languages = append(languages, lang)
	}
	return languages
}
