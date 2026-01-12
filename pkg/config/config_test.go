package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_Basic(t *testing.T) {
	cfg := New()

	// 测试 Set 和 Get
	cfg.Set("name", "test")
	cfg.Set("port", 8080)
	cfg.Set("enabled", true)

	if val := cfg.GetString("name"); val != "test" {
		t.Errorf("expected 'test', got '%s'", val)
	}

	if val := cfg.GetInt("port"); val != 8080 {
		t.Errorf("expected 8080, got %d", val)
	}

	if val := cfg.GetBool("enabled"); val != true {
		t.Errorf("expected true, got %v", val)
	}
}

func TestConfig_NestedAccess(t *testing.T) {
	cfg := New()

	// 测试嵌套路径
	cfg.Set("database.host", "localhost")
	cfg.Set("database.port", 3306)
	cfg.Set("database.credentials.username", "admin")
	cfg.Set("database.credentials.password", "secret")

	if val := cfg.GetString("database.host"); val != "localhost" {
		t.Errorf("expected 'localhost', got '%s'", val)
	}

	if val := cfg.GetInt("database.port"); val != 3306 {
		t.Errorf("expected 3306, got %d", val)
	}

	if val := cfg.GetString("database.credentials.username"); val != "admin" {
		t.Errorf("expected 'admin', got '%s'", val)
	}
}

func TestConfig_GetSection(t *testing.T) {
	cfg := New()
	cfg.Set("database.host", "localhost")
	cfg.Set("database.port", 3306)

	section := cfg.GetSection("database")
	if len(section) != 2 {
		t.Errorf("expected 2 items in section, got %d", len(section))
	}

	if section["host"] != "localhost" {
		t.Errorf("expected 'localhost', got '%v'", section["host"])
	}
}

func TestConfig_Unmarshal(t *testing.T) {
	type AppConfig struct {
		Name    string
		Port    int
		Enabled bool
	}

	cfg := New()
	cfg.Set("Name", "MyApp")
	cfg.Set("Port", 8080)
	cfg.Set("Enabled", true)

	var app AppConfig
	if err := cfg.Unmarshal(&app); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if app.Name != "MyApp" {
		t.Errorf("expected 'MyApp', got '%s'", app.Name)
	}

	if app.Port != 8080 {
		t.Errorf("expected 8080, got %d", app.Port)
	}

	if app.Enabled != true {
		t.Errorf("expected true, got %v", app.Enabled)
	}
}

func TestConfig_UnmarshalKey(t *testing.T) {
	type Database struct {
		Host string
		Port int
	}

	cfg := New()
	cfg.Set("database.Host", "localhost")
	cfg.Set("database.Port", 3306)

	var db Database
	if err := cfg.UnmarshalKey("database", &db); err != nil {
		t.Fatalf("failed to unmarshal key: %v", err)
	}

	if db.Host != "localhost" {
		t.Errorf("expected 'localhost', got '%s'", db.Host)
	}

	if db.Port != 3306 {
		t.Errorf("expected 3306, got %d", db.Port)
	}
}

func TestParser_JSON(t *testing.T) {
	jsonData := []byte(`{"name": "test", "port": 8080, "enabled": true}`)

	result, err := parse(jsonData, FormatJSON)
	if err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("expected 'test', got '%v'", result["name"])
	}

	if result["port"].(float64) != 8080 {
		t.Errorf("expected 8080, got %v", result["port"])
	}

	if result["enabled"] != true {
		t.Errorf("expected true, got %v", result["enabled"])
	}
}

func TestParser_YAML(t *testing.T) {
	yamlData := []byte(`
name: test
port: 8080
enabled: true
`)

	result, err := parse(yamlData, FormatYAML)
	if err != nil {
		t.Fatalf("failed to parse YAML: %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("expected 'test', got '%v'", result["name"])
	}

	if result["port"].(int) != 8080 {
		t.Errorf("expected 8080, got %v", result["port"])
	}

	if result["enabled"] != true {
		t.Errorf("expected true, got %v", result["enabled"])
	}
}

func TestParser_TOML(t *testing.T) {
	tomlData := []byte(`
name = "test"
port = 8080
enabled = true
`)

	result, err := parse(tomlData, FormatTOML)
	if err != nil {
		t.Fatalf("failed to parse TOML: %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("expected 'test', got '%v'", result["name"])
	}

	if result["port"].(int64) != 8080 {
		t.Errorf("expected 8080, got %v", result["port"])
	}

	if result["enabled"] != true {
		t.Errorf("expected true, got %v", result["enabled"])
	}
}

func TestLoad(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建测试配置文件
	jsonFile := filepath.Join(tempDir, "config.json")
	jsonData := `{"name": "test", "port": 8080}`
	if err := os.WriteFile(jsonFile, []byte(jsonData), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// 加载配置
	cfg, err := Load(jsonFile)
	if err != nil {
		t.Fatalf("failed to load file: %v", err)
	}

	if val := cfg.GetString("name"); val != "test" {
		t.Errorf("expected 'test', got '%s'", val)
	}
}

func TestLoad_EnvVars(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("TEST_HOST", "localhost")
	os.Setenv("TEST_PORT", "3306")
	defer os.Unsetenv("TEST_HOST")
	defer os.Unsetenv("TEST_PORT")

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")
	configData := `{
		"host": "${TEST_HOST}",
		"port": "${TEST_PORT}",
		"default": "${NONEXISTENT:default_value}"
	}`

	if err := os.WriteFile(configFile, []byte(configData), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("failed to load config with env: %v", err)
	}

	if val := cfg.GetString("host"); val != "localhost" {
		t.Errorf("expected 'localhost', got '%s'", val)
	}

	if val := cfg.GetString("port"); val != "3306" {
		t.Errorf("expected '3306', got '%s'", val)
	}

	if val := cfg.GetString("default"); val != "default_value" {
		t.Errorf("expected 'default_value', got '%s'", val)
	}
}

func TestConfig_GetWithDefault(t *testing.T) {
	cfg := New()
	cfg.Set("existing", "value")

	// 测试存在的key
	if val := cfg.GetWithDefault("existing", "default").(string); val != "value" {
		t.Errorf("expected 'value', got '%s'", val)
	}

	// 测试不存在的key
	if val := cfg.GetWithDefault("nonexistent", "default").(string); val != "default" {
		t.Errorf("expected 'default', got '%s'", val)
	}
}

func TestConfig_Has(t *testing.T) {
	cfg := New()
	cfg.Set("existing", "value")

	if !cfg.Has("existing") {
		t.Error("expected key to exist")
	}

	if cfg.Has("nonexistent") {
		t.Error("expected key not to exist")
	}
}

func TestLoadFromBytes(t *testing.T) {
	jsonData := []byte(`{"name": "test", "port": 8080}`)

	cfg, err := LoadFromBytes(jsonData, FormatJSON)
	if err != nil {
		t.Fatalf("failed to load from bytes: %v", err)
	}

	if val := cfg.GetString("name"); val != "test" {
		t.Errorf("expected 'test', got '%s'", val)
	}
}

// 测试类型安全的默认值方法
func TestGetWithDefaultTyped(t *testing.T) {
	cfg := New()
	cfg.Set("name", "test")
	cfg.Set("port", 8080)
	cfg.Set("enabled", true)
	cfg.Set("ratio", 3.14)

	// 测试存在的值
	if val := cfg.GetStringWithDefault("name", "default"); val != "test" {
		t.Errorf("expected 'test', got '%s'", val)
	}

	if val := cfg.GetIntWithDefault("port", 9090); val != 8080 {
		t.Errorf("expected 8080, got %d", val)
	}

	if val := cfg.GetBoolWithDefault("enabled", false); val != true {
		t.Errorf("expected true, got %v", val)
	}

	if val := cfg.GetFloatWithDefault("ratio", 1.0); val != 3.14 {
		t.Errorf("expected 3.14, got %f", val)
	}

	// 测试不存在的值（返回默认值）
	if val := cfg.GetStringWithDefault("missing", "default"); val != "default" {
		t.Errorf("expected 'default', got '%s'", val)
	}

	if val := cfg.GetIntWithDefault("missing", 9090); val != 9090 {
		t.Errorf("expected 9090, got %d", val)
	}

	if val := cfg.GetBoolWithDefault("missing", true); val != true {
		t.Errorf("expected true, got %v", val)
	}

	if val := cfg.GetFloatWithDefault("missing", 2.5); val != 2.5 {
		t.Errorf("expected 2.5, got %f", val)
	}
}

// 测试切片支持
func TestGetSlice(t *testing.T) {
	cfg := New()
	cfg.Set("hosts", []any{"host1", "host2", "host3"})
	cfg.Set("ports", []any{8080, 8081, 8082})
	cfg.Set("mixed", []any{"string", 123, true})

	// 测试字符串切片
	hosts := cfg.GetStringSlice("hosts")
	if len(hosts) != 3 {
		t.Errorf("expected 3 hosts, got %d", len(hosts))
	}
	if hosts[0] != "host1" || hosts[1] != "host2" || hosts[2] != "host3" {
		t.Errorf("unexpected hosts: %v", hosts)
	}

	// 测试整数切片
	ports := cfg.GetIntSlice("ports")
	if len(ports) != 3 {
		t.Errorf("expected 3 ports, got %d", len(ports))
	}
	if ports[0] != 8080 || ports[1] != 8081 || ports[2] != 8082 {
		t.Errorf("unexpected ports: %v", ports)
	}

	// 测试任意类型切片
	mixed := cfg.GetSlice("mixed")
	if len(mixed) != 3 {
		t.Errorf("expected 3 items, got %d", len(mixed))
	}

	// 测试不存在的切片
	empty := cfg.GetStringSlice("nonexistent")
	if len(empty) != 0 {
		t.Errorf("expected empty slice, got %v", empty)
	}
}

// 测试 Load 方法的替换行为
func TestConfigLoadReplace(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "config1.json")
	data1 := []byte(`{"app": {"name": "test", "version": "1.0"}, "port": 8080}`)
	if err := os.WriteFile(file1, data1, 0644); err != nil {
		t.Fatal(err)
	}

	file2 := filepath.Join(tmpDir, "config2.json")
	data2 := []byte(`{"app": {"version": "2.0", "debug": true}, "timeout": 30}`)
	if err := os.WriteFile(file2, data2, 0644); err != nil {
		t.Fatal(err)
	}

	cfg := New()

	// 加载第一个文件
	if err := cfg.Load(file1); err != nil {
		t.Fatal(err)
	}

	// 验证第一个文件的内容
	if val := cfg.GetString("app.name"); val != "test" {
		t.Errorf("expected 'test', got '%s'", val)
	}
	if val := cfg.GetString("app.version"); val != "1.0" {
		t.Errorf("expected '1.0', got '%s'", val)
	}
	if val := cfg.GetInt("port"); val != 8080 {
		t.Errorf("expected 8080, got %d", val)
	}

	// 加载第二个文件（应该完全替换）
	if err := cfg.Load(file2); err != nil {
		t.Fatal(err)
	}

	// 验证替换结果：第一个文件的配置应该被完全替换
	if cfg.Has("app.name") {
		t.Error("app.name should not exist after replacement")
	}
	if val := cfg.GetString("app.version"); val != "2.0" {
		t.Errorf("expected '2.0', got '%s'", val)
	}
	if val := cfg.GetBool("app.debug"); val != true {
		t.Errorf("expected true, got %v", val)
	}
	if cfg.Has("port") {
		t.Error("port should not exist after replacement")
	}
	if val := cfg.GetInt("timeout"); val != 30 {
		t.Errorf("expected 30, got %d", val)
	}
}

// 测试 Load 方法在空 Config 时的行为（直接加载）
func TestConfigLoadOnEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "config1.json")
	data1 := []byte(`{"app": "test1", "port": 8080}`)
	if err := os.WriteFile(file1, data1, 0644); err != nil {
		t.Fatal(err)
	}

	file2 := filepath.Join(tmpDir, "config2.json")
	data2 := []byte(`{"app": "test2", "timeout": 30}`)
	if err := os.WriteFile(file2, data2, 0644); err != nil {
		t.Fatal(err)
	}

	cfg := New()

	// 加载第一个文件（空 Config，应该直接加载）
	if err := cfg.Load(file1); err != nil {
		t.Fatal(err)
	}

	// 验证加载结果
	if val := cfg.GetString("app"); val != "test1" {
		t.Errorf("expected 'test1', got '%s'", val)
	}
	if val := cfg.GetInt("port"); val != 8080 {
		t.Errorf("expected 8080, got %d", val)
	}

	// Load 本身就是完全替换，但这里先 Clear 再 Load 来演示替换行为
	cfg.Clear()
	if err := cfg.Load(file2); err != nil {
		t.Fatal(err)
	}

	// 验证替换结果
	if val := cfg.GetString("app"); val != "test2" {
		t.Errorf("expected 'test2', got '%s'", val)
	}
	if val := cfg.GetInt("timeout"); val != 30 {
		t.Errorf("expected 30, got %d", val)
	}
	// port 应该不存在了
	if cfg.Has("port") {
		t.Error("port should not exist after Clear and Load")
	}
}

// 测试包级函数返回独立实例
func TestLoadIndependentInstances(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "config1.json")
	data1 := []byte(`{"name": "config1"}`)
	if err := os.WriteFile(file1, data1, 0644); err != nil {
		t.Fatal(err)
	}

	file2 := filepath.Join(tmpDir, "config2.json")
	data2 := []byte(`{"name": "config2"}`)
	if err := os.WriteFile(file2, data2, 0644); err != nil {
		t.Fatal(err)
	}

	// 加载两个配置
	cfg1, err := Load(file1)
	if err != nil {
		t.Fatal(err)
	}

	cfg2, err := Load(file2)
	if err != nil {
		t.Fatal(err)
	}

	// 验证它们是不同的实例
	if cfg1 == cfg2 {
		t.Error("Load should return different Config instances")
	}

	// 验证各自的值
	if val := cfg1.GetString("name"); val != "config1" {
		t.Errorf("cfg1 expected 'config1', got '%s'", val)
	}

	if val := cfg2.GetString("name"); val != "config2" {
		t.Errorf("cfg2 expected 'config2', got '%s'", val)
	}
}

// 测试 MustLoad 方法
func TestMustLoad(t *testing.T) {
	tmpDir := t.TempDir()

	// 测试成功的情况
	file := filepath.Join(tmpDir, "config.json")
	data := []byte(`{"name": "test", "port": 8080}`)
	if err := os.WriteFile(file, data, 0644); err != nil {
		t.Fatal(err)
	}

	cfg := MustLoad(file)
	if val := cfg.GetString("name"); val != "test" {
		t.Errorf("expected 'test', got '%s'", val)
	}
}

// 测试 MustUnmarshal 方法
func TestMustUnmarshal(t *testing.T) {
	tmpDir := t.TempDir()

	file := filepath.Join(tmpDir, "config.json")
	// JSON 键名需要与结构体字段名匹配（大小写一致）
	data := []byte(`{"Name": "test-app", "Port": 8080, "Enabled": true}`)
	if err := os.WriteFile(file, data, 0644); err != nil {
		t.Fatal(err)
	}

	type AppConfig struct {
		Name    string
		Port    int
		Enabled bool
	}

	var c AppConfig
	MustUnmarshal(file, &c)

	if c.Name != "test-app" {
		t.Errorf("expected 'test-app', got '%s'", c.Name)
	}
	if c.Port != 8080 {
		t.Errorf("expected 8080, got %d", c.Port)
	}
	if !c.Enabled {
		t.Error("expected true")
	}
}

// 测试 MustUnmarshal 方法（支持环境变量）
func TestMustUnmarshal_EnvVars(t *testing.T) {
	tmpDir := t.TempDir()

	// 设置环境变量
	os.Setenv("TEST_APP_NAME", "my-app")
	os.Setenv("TEST_PORT", "9090")
	defer os.Unsetenv("TEST_APP_NAME")
	defer os.Unsetenv("TEST_PORT")

	file := filepath.Join(tmpDir, "config.json")
	// JSON 键名需要与结构体字段名匹配（大小写一致）
	data := []byte(`{
		"Name": "${TEST_APP_NAME}",
		"Port": "${TEST_PORT}"
	}`)
	if err := os.WriteFile(file, data, 0644); err != nil {
		t.Fatal(err)
	}

	type AppConfig struct {
		Name string
		Port string
	}

	var c AppConfig
	MustUnmarshal(file, &c)

	if c.Name != "my-app" {
		t.Errorf("expected 'my-app', got '%s'", c.Name)
	}
	if c.Port != "9090" {
		t.Errorf("expected '9090', got '%s'", c.Port)
	}
}

// 测试 WriteToFile 方法
func TestWriteToFile(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建配置
	cfg := New()
	cfg.Set("app.name", "TestApp")
	cfg.Set("app.port", 8080)
	cfg.Set("database.host", "localhost")
	cfg.Set("database.port", 3306)

	// 测试导出为 JSON
	jsonFile := filepath.Join(tmpDir, "output.json")
	if err := cfg.WriteToFile(jsonFile); err != nil {
		t.Fatalf("failed to write JSON: %v", err)
	}

	// 验证 JSON 文件内容
	loadedJSON, err := Load(jsonFile)
	if err != nil {
		t.Fatalf("failed to load JSON: %v", err)
	}
	if loadedJSON.GetString("app.name") != "TestApp" {
		t.Error("JSON content mismatch")
	}

	// 测试导出为 YAML
	yamlFile := filepath.Join(tmpDir, "output.yaml")
	if err := cfg.WriteToFile(yamlFile); err != nil {
		t.Fatalf("failed to write YAML: %v", err)
	}

	// 验证 YAML 文件内容
	loadedYAML, err := Load(yamlFile)
	if err != nil {
		t.Fatalf("failed to load YAML: %v", err)
	}
	if loadedYAML.GetInt("app.port") != 8080 {
		t.Error("YAML content mismatch")
	}

	// 测试导出为 TOML
	tomlFile := filepath.Join(tmpDir, "output.toml")
	if err := cfg.WriteToFile(tomlFile); err != nil {
		t.Fatalf("failed to write TOML: %v", err)
	}

	// 验证 TOML 文件内容
	loadedTOML, err := Load(tomlFile)
	if err != nil {
		t.Fatalf("failed to load TOML: %v", err)
	}
	if loadedTOML.GetString("database.host") != "localhost" {
		t.Error("TOML content mismatch")
	}
}

// 测试 WriteToFile 的实际使用场景
func TestWriteToFile_UseCases(t *testing.T) {
	tmpDir := t.TempDir()

	// 场景1: 配置替换后导出检查
	base := filepath.Join(tmpDir, "base.json")
	os.WriteFile(base, []byte(`{"name": "app", "port": 8080}`), 0644)

	prod := filepath.Join(tmpDir, "prod.json")
	os.WriteFile(prod, []byte(`{"port": 9090, "env": "production"}`), 0644)

	cfg := New()
	cfg.Load(base)

	// 验证第一个文件的内容
	if cfg.GetString("name") != "app" {
		t.Error("name should be 'app'")
	}
	if cfg.GetInt("port") != 8080 {
		t.Error("port should be 8080")
	}

	// 加载第二个文件（应该完全替换）
	cfg.Load(prod)

	exported := filepath.Join(tmpDir, "exported.json")
	if err := cfg.WriteToFile(exported); err != nil {
		t.Fatalf("failed to write exported config: %v", err)
	}

	// 验证替换结果
	result, _ := Load(exported)
	if result.Has("name") {
		t.Error("name should not exist after replacement")
	}
	if result.GetInt("port") != 9090 {
		t.Error("port should be 9090")
	}
	if result.GetString("env") != "production" {
		t.Error("env should be 'production'")
	}

	// 场景2: 动态修改后保存
	cfg.Set("debug", true)
	cfg.Set("new_field", "added")

	updated := filepath.Join(tmpDir, "updated.yaml")
	if err := cfg.WriteToFile(updated); err != nil {
		t.Fatalf("failed to write updated config: %v", err)
	}

	// 验证修改是否保存
	result2, _ := Load(updated)
	if !result2.GetBool("debug") {
		t.Error("debug should be true")
	}
	if result2.GetString("new_field") != "added" {
		t.Error("new_field should be 'added'")
	}
}
