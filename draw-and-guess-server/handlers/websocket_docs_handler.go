package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// GetAsyncAPISpec godoc
// @Summary AsyncAPI 스펙 조회
// @Description WebSocket API의 AsyncAPI 3.0 스펙을 반환합니다 (Accept 헤더에 따라 JSON 또는 YAML)
// @Tags WebSocket
// @Produce json
// @Produce yaml
// @Success 200 {object} map[string]interface{} "AsyncAPI 스펙"
// @Router /asyncapi.yaml [get]
// @Router /asyncapi.json [get]
func GetAsyncAPISpec(c *gin.Context) {
	// AsyncAPI YAML 파일 읽기
	asyncAPIPath := filepath.Join("docs", "asyncapi.yaml")
	data, err := os.ReadFile(asyncAPIPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read AsyncAPI spec file",
		})
		return
	}

	// Accept 헤더에 따라 응답 형식 결정
	accept := c.GetHeader("Accept")

	if accept == "application/json" {
		// YAML을 JSON으로 변환
		var spec map[string]interface{}
		if err := yaml.Unmarshal(data, &spec); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to parse AsyncAPI spec",
			})
			return
		}
		c.JSON(http.StatusOK, spec)
	} else {
		// YAML 그대로 반환
		c.Data(http.StatusOK, "application/x-yaml", data)
	}
}

// AsyncAPIDocumentation godoc
// @Summary AsyncAPI HTML 문서
// @Description WebSocket API의 AsyncAPI HTML 문서를 제공합니다
// @Tags WebSocket
// @Produce html
// @Success 200 {string} string "HTML 문서"
// @Router /asyncapi [get]
func AsyncAPIDocumentation(c *gin.Context) {
	// 쿼리 파라미터로 뷰어 선택 가능: ?viewer=playground|studio|react|web
	viewer := c.DefaultQuery("viewer", "playground")

	var html string

	switch viewer {
	case "studio":
		// AsyncAPI Studio (가장 풍부한 기능)
		html = getStudioHTML()
	case "react":
		// AsyncAPI React Component (커스터마이징 가능)
		html = getReactComponentHTML()
	case "web":
		// AsyncAPI Web Component (심플)
		html = getWebComponentHTML()
	default:
		// AsyncAPI Playground (추천 - 인터랙티브)
		html = getPlaygroundHTML()
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// AsyncAPI Playground - 가장 모던하고 인터랙티브
func getPlaygroundHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Draw and Guess Game - WebSocket API Documentation</title>
    <style>
        body {
            margin: 0;
            padding: 0;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
        }
        .container {
            width: 100%;
            height: 100vh;
        }
        .header {
            background: #1a1a1a;
            color: white;
            padding: 15px 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .viewer-select {
            padding: 8px 12px;
            border-radius: 4px;
            background: #333;
            color: white;
            border: none;
            cursor: pointer;
        }
        iframe {
            width: 100%;
            height: calc(100vh - 60px);
            border: none;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>🎨 Draw and Guess - WebSocket API</h2>
            <select class="viewer-select" onchange="location.href='?viewer='+this.value">
                <option value="playground" selected>Playground (추천)</option>
                <option value="studio">Studio (풀 기능)</option>
                <option value="react">React Component</option>
                <option value="web">Web Component</option>
            </select>
        </div>
        <iframe 
            src="https://playground.asyncapi.io/?url=` + getBaseURL() + `/app/asyncapi.yaml"
            title="AsyncAPI Playground">
        </iframe>
    </div>
</body>
</html>`
}

// AsyncAPI Studio - 가장 풍부한 기능 (편집, 검증, 내보내기 등)
func getStudioHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Draw and Guess Game - WebSocket API Documentation</title>
    <style>
        body { margin: 0; padding: 0; font-family: sans-serif; }
        .header {
            background: #1a1a1a;
            color: white;
            padding: 15px 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .viewer-select {
            padding: 8px 12px;
            border-radius: 4px;
            background: #333;
            color: white;
            border: none;
            cursor: pointer;
        }
        iframe { width: 100%; height: calc(100vh - 60px); border: none; }
    </style>
</head>
<body>
    <div class="header">
        <h2>🎨 Draw and Guess - WebSocket API (Studio)</h2>
        <select class="viewer-select" onchange="location.href='?viewer='+this.value">
            <option value="playground">Playground (추천)</option>
            <option value="studio" selected>Studio (풀 기능)</option>
            <option value="react">React Component</option>
            <option value="web">Web Component</option>
        </select>
    </div>
    <iframe 
        src="https://studio.asyncapi.com/?url=` + getBaseURL() + `/app/asyncapi.yaml"
        title="AsyncAPI Studio">
    </iframe>
</body>
</html>`
}

// AsyncAPI React Component - 커스터마이징 가능
func getReactComponentHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Draw and Guess Game - WebSocket API Documentation</title>
    <link rel="stylesheet" href="https://unpkg.com/@asyncapi/react-component@1.4.11/styles/default.min.css">
    <style>
        body { margin: 0; padding: 0; font-family: sans-serif; }
        .header {
            background: #1a1a1a;
            color: white;
            padding: 15px 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .viewer-select {
            padding: 8px 12px;
            border-radius: 4px;
            background: #333;
            color: white;
            border: none;
            cursor: pointer;
        }
        #asyncapi { padding: 20px; }
    </style>
</head>
<body>
    <div class="header">
        <h2>🎨 Draw and Guess - WebSocket API (React)</h2>
        <select class="viewer-select" onchange="location.href='?viewer='+this.value">
            <option value="playground">Playground (추천)</option>
            <option value="studio">Studio (풀 기능)</option>
            <option value="react" selected>React Component</option>
            <option value="web">Web Component</option>
        </select>
    </div>
    <div id="asyncapi"></div>
    
    <script src="https://unpkg.com/@asyncapi/react-component@1.4.11/browser/standalone/index.js"></script>
    <script>
        fetch('/app/asyncapi.yaml')
            .then(response => response.text())
            .then(schema => {
                AsyncApiStandalone.render({
                    schema: schema,
                    config: {
                        show: {
                            sidebar: true,
                            info: true,
                            servers: true,
                            operations: true,
                            messages: true,
                            schemas: true,
                        },
                        sidebar: {
                            showOperations: 'byDefault',
                        },
                    },
                }, document.getElementById('asyncapi'));
            });
    </script>
</body>
</html>`
}

// AsyncAPI Web Component - 심플한 버전
func getWebComponentHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Draw and Guess Game - WebSocket API Documentation</title>
    <style>
        body { margin: 0; padding: 0; font-family: sans-serif; }
        .header {
            background: #1a1a1a;
            color: white;
            padding: 15px 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .viewer-select {
            padding: 8px 12px;
            border-radius: 4px;
            background: #333;
            color: white;
            border: none;
            cursor: pointer;
        }
        asyncapi-component {
            display: block;
            padding: 20px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h2>🎨 Draw and Guess - WebSocket API (Web Component)</h2>
        <select class="viewer-select" onchange="location.href='?viewer='+this.value">
            <option value="playground">Playground (추천)</option>
            <option value="studio">Studio (풀 기능)</option>
            <option value="react">React Component</option>
            <option value="web" selected>Web Component</option>
        </select>
    </div>
    <asyncapi-component
        schemaUrl="/app/asyncapi.yaml"
        schemaFetchOptions='{"method":"GET","mode":"cors"}'>
    </asyncapi-component>
    
    <script src="https://unpkg.com/@asyncapi/web-component@latest/lib/asyncapi-web-component.js" defer></script>
</body>
</html>`
}

func getBaseURL() string {
	// 프로덕션 환경에서는 실제 도메인 사용
	return "http://localhost:8080"
}
