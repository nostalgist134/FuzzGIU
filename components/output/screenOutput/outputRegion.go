package output

import (
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// init 初始化输出区域
func (r *screenOutputRegion) init(maxLines int, noBorder ...bool) {
	r.Pg = widgets.NewParagraph()
	r.Pg.WrapText = false
	r.lineLeft = 0
	r.lines = make([]string, 0)
	r.maxRenderLines = maxLines
	r.renderBuffer = make([]string, maxLines)
	if len(noBorder) > 0 && noBorder[0] {
		r.Pg.Border = false
	}
}

func (r *screenOutputRegion) clearRenderBuffer() {
	for i := 0; i < len(r.renderBuffer); i++ {
		r.renderBuffer[i] = ""
	}
}

// render 渲染outputRegion对象，如果title不为空，则渲染标题
func (r *screenOutputRegion) render(title string, unlock ...bool) {
	if len(unlock) == 0 || !unlock[0] {
		r.mu.Lock()
		defer r.mu.Unlock()
	}
	if !r.rendered {
		r.rendered = true
		r.Pg.SetRect(r.TopCorner.X, r.TopCorner.Y, r.BottomCorner.X, r.BottomCorner.Y)
	}
	// 设置标题
	if title != "" {
		r.Pg.Title = title
	}
	// 将要渲染的各行按照最大长度截断后再渲染
	r.clearRenderBuffer()
	if truncateLines(r.renderBuffer, r.lines, r.lineInd, r.lineLeft, r.maxRenderLines,
		r.BottomCorner.X-r.TopCorner.X+10) {
		r.lineLeft--
		truncateLines(r.renderBuffer, r.lines, r.lineInd, r.lineLeft, r.maxRenderLines,
			r.BottomCorner.X-r.TopCorner.X+10)
	}
	r.Pg.Text = lines2Text(r.renderBuffer)
	screenOutput.renderMu.Lock()
	defer screenOutput.renderMu.Unlock()
	if !outputHasInit.Load() {
		return
	}
	ui.Render(r.Pg)
}

// clear 将outputRegion清空
func (r *screenOutputRegion) clear() {
	r.lines = r.lines[:0]
	r.lineLeft = 0
	r.lineLeft = 0
	r.render("")
}

// setRect 设置渲染段落的对角
func (r *screenOutputRegion) setRect(pos []int) {
	if len(pos) < 4 {
		return
	}
	setP := func(p *int, v int) {
		if v > 0 {
			*p = v
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	setP(&r.TopCorner.X, pos[0])
	setP(&r.TopCorner.Y, pos[1])
	setP(&r.BottomCorner.X, pos[2])
	setP(&r.BottomCorner.Y, pos[3])
	r.Pg.SetRect(r.TopCorner.X, r.TopCorner.Y, r.BottomCorner.X, r.BottomCorner.Y)
}

func (r *screenOutputRegion) append(lines []string) {
	r.lines = append(r.lines, lines...)
}

func (r *screenOutputRegion) setLines(lines []string) {
	r.lines = lines
}

// scroll 用来控制窗口的上下左右翻页
func (r *screenOutputRegion) scroll(direction int8) {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch direction {
	case directionUp:
		if r.lineInd > 0 {
			r.lineInd--
			r.render("", true)
		}
	case directionDown:
		if r.lineInd < len(r.lines)-1 {
			r.lineInd++
			r.render("", true)
		}
	case directionLeft:
		if r.lineLeft > 0 {
			r.lineLeft--
			r.render("", true)
		}
	case directionRight:
		r.lineLeft++
		r.render("", true)
	}
}

// switchHighLightRegion 切换高亮显示的窗口
func switchHighLightRegion(lastInd int) {
	selectableRegions[indSelect].mu.Lock()
	defer selectableRegions[indSelect].mu.Unlock()
	selectableRegions[indSelect].Pg.BorderStyle.Fg = ui.ColorCyan
	ui.Render(selectableRegions[indSelect].Pg)
	// 将原先的高亮窗口解除
	if lastInd >= 0 && lastInd < len(selectableRegions) && lastInd != indSelect {
		selectableRegions[lastInd].mu.Lock()
		defer selectableRegions[lastInd].mu.Unlock()
		selectableRegions[lastInd].Pg.BorderStyle.Fg = ui.ColorWhite
		ui.Render(selectableRegions[lastInd].Pg)
	}
}
