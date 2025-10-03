package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sync"

	"github.com/b25/strategy-engine/internal/strategies"
	"github.com/b25/strategy-engine/pkg/logger"
	"github.com/b25/strategy-engine/pkg/metrics"
)

// PluginLoader handles loading and reloading of strategy plugins
type PluginLoader struct {
	pluginsDir  string
	logger      *logger.Logger
	metrics     *metrics.Collector
	loadedPlugins map[string]*plugin.Plugin
	mu          sync.RWMutex
}

// NewPluginLoader creates a new plugin loader
func NewPluginLoader(pluginsDir string, log *logger.Logger, m *metrics.Collector) *PluginLoader {
	return &PluginLoader{
		pluginsDir:    pluginsDir,
		logger:        log,
		metrics:       m,
		loadedPlugins: make(map[string]*plugin.Plugin),
	}
}

// Load loads all plugins from the plugins directory
func (p *PluginLoader) Load() error {
	p.logger.Info("Loading plugins from directory", "dir", p.pluginsDir)

	// Check if directory exists
	if _, err := os.Stat(p.pluginsDir); os.IsNotExist(err) {
		p.logger.Warn("Plugins directory does not exist", "dir", p.pluginsDir)
		return nil
	}

	// Find all .so files
	pattern := filepath.Join(p.pluginsDir, "*.so")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to list plugin files: %w", err)
	}

	p.logger.Info("Found plugin files", "count", len(files))

	for _, file := range files {
		if err := p.loadPlugin(file); err != nil {
			p.logger.Error("Failed to load plugin",
				"file", file,
				"error", err,
			)
			continue
		}
	}

	return nil
}

// loadPlugin loads a single plugin
func (p *PluginLoader) loadPlugin(path string) error {
	p.logger.Info("Loading plugin", "path", path)

	// Open the plugin
	plug, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	// Look for the NewStrategy function
	symbol, err := plug.Lookup("NewStrategy")
	if err != nil {
		return fmt.Errorf("plugin missing NewStrategy function: %w", err)
	}

	// Type assert to the expected function signature
	newStrategy, ok := symbol.(func() strategies.Strategy)
	if !ok {
		return fmt.Errorf("invalid NewStrategy function signature")
	}

	// Create a test instance to get the strategy name
	testStrategy := newStrategy()
	strategyName := testStrategy.Name()

	p.logger.Info("Loaded plugin strategy",
		"name", strategyName,
		"path", path,
	)

	p.mu.Lock()
	p.loadedPlugins[path] = plug
	p.mu.Unlock()

	return nil
}

// Reload reloads all plugins
func (p *PluginLoader) Reload() error {
	p.logger.Info("Reloading plugins")

	// Note: Go plugins cannot be unloaded, so we can only load new ones
	// In a production system, you'd need to restart the process for true reload
	return p.Load()
}

// PythonPluginRunner handles Python-based strategies
type PythonPluginRunner struct {
	scriptPath string
	pythonPath string
	venvPath   string
	logger     *logger.Logger
}

// NewPythonPluginRunner creates a new Python plugin runner
func NewPythonPluginRunner(scriptPath, pythonPath, venvPath string, log *logger.Logger) *PythonPluginRunner {
	return &PythonPluginRunner{
		scriptPath: scriptPath,
		pythonPath: pythonPath,
		venvPath:   venvPath,
		logger:     log,
	}
}

// Run executes a Python strategy script
// This is a placeholder - in production, you'd use a proper IPC mechanism
func (r *PythonPluginRunner) Run() error {
	r.logger.Info("Python plugin support is a placeholder",
		"script", r.scriptPath,
	)

	// In production, you would:
	// 1. Start a Python process
	// 2. Establish IPC (gRPC, ZeroMQ, or named pipes)
	// 3. Send market data and receive signals
	// 4. Handle the lifecycle

	return nil
}
