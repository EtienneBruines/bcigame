package systems

import (
	"image"
	"log"
	"strconv"

	"github.com/EtienneBruines/gobci"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg/draw"
	"github.com/gonum/plot/vg/vgimg"
	"github.com/paked/engi"
)

type Calibrate struct {
	*engi.System

	Connection *gobci.Connection
	Header     *gobci.Header

	channels []channelXYer
}

func (Calibrate) Type() string {
	return "CalibrateSystem"
}

func (c *Calibrate) New() {
	c.System = engi.NewSystem()

	var err error
	c.Connection, err = gobci.Connect("")
	if err != nil {
		log.Fatal(err)
	}

	err = c.Connection.FlushData()
	if err != nil {
		log.Fatal(err)
	}

	// Get latest header info
	c.Header, err = c.Connection.GetHeader()
	if err != nil {
		log.Fatal(err)
	}

	engi.Mailbox.Listen("CalibrateMessage", func(m engi.Message) {
		cm, ok := m.(CalibrateMessage)
		if !ok {
			return
		}

		if cm.Enable {
			c.drawScene()
		} else {
			c.destroyScene()
		}
	})
}

func (c *Calibrate) destroyScene() {
	engi.Mailbox.Dispatch(engi.PauseMessage{false})
}

func (c *Calibrate) drawScene() {
	engi.Mailbox.Dispatch(engi.PauseMessage{true})

	for i := uint32(0); i < c.Header.NChannels; i++ {
		e := engi.NewEntity([]string{c.Type(), "RenderSystem"})
		espace := &engi.SpaceComponent{engi.Point{0, float32(i * (3*dpi + 10))}, 0, 0}
		e.AddComponent(espace)
		e.AddComponent(&engi.UnpauseComponent{})
		e.AddComponent(&CalibrateComponent{i})

		c.World.AddEntity(e)
	}
}

func (c *Calibrate) Pre() {
	var err error

	c.Header.NSamples, c.Header.NEvents, err = c.Connection.WaitData(0, 0, 0)
	if err != nil {
		log.Fatal(err)
	}

	// Get actual data
	min := uint32(0)
	if c.Header.NSamples >= uint32(c.Header.SamplingFrequency*timePeriod) {
		min = c.Header.NSamples - uint32(c.Header.SamplingFrequency*timePeriod)
	}
	max := uint32(0)
	if c.Header.NSamples > 0 {
		max = c.Header.NSamples - 1
	}

	samples, err := c.Connection.GetData(min, max)
	if err != nil {
		log.Fatal(err)
	}

	// Visualizing the channels
	c.channels = make([]channelXYer, c.Header.NChannels)
	for _, sample := range samples {
		for i := uint32(0); i < c.Header.NChannels; i++ {
			c.channels[i].Values = append(c.channels[i].Values, sample[i])
		}
	}

	for chIndex := range c.channels {
		c.channels[chIndex].freq = c.Header.SamplingFrequency
	}
}

func (c *Calibrate) Update(entity *engi.Entity, dt float32) {
	var cal *CalibrateComponent
	if !entity.GetComponent(&cal) {
		return
	}

	// Render the image again
	plt, err := plot.New()
	if err != nil {
		log.Fatal(err)
	}

	plotutil.AddLinePoints(plt,
		"CH"+strconv.Itoa(int(cal.ChannelIndex)), plotter.XYer(c.channels[cal.ChannelIndex]))
	img := image.NewRGBA(image.Rect(0, 0, 3*dpi, 3*dpi))
	canv := vgimg.NewWith(vgimg.UseImage(img))
	plt.Draw(draw.New(canv))
	bgTexture := engi.NewImageRGBA(img)

	// Give it to engi

	erender := &engi.RenderComponent{
		Display:      engi.NewRegion(engi.NewTexture(bgTexture), 0, 0, 3*dpi, 3*dpi),
		Scale:        engi.Point{1, 1},
		Transparency: 1,
		Color:        0xffffff,
	}
	erender.SetPriority(engi.HUDGround)

	entity.AddComponent(erender)
}

type CalibrateComponent struct {
	ChannelIndex uint32
}

func (CalibrateComponent) Type() string {
	return "CalibrateComponent"
}

type CalibrateMessage struct {
	Enable bool
}

func (CalibrateMessage) Type() string {
	return "CalibrateMessage"
}

var timePeriod = float32(5) // seconds
const dpi = 96

type channelXYer struct {
	Values []float64
	freq   float32
}

func (c channelXYer) Len() int {
	if max := c.freq * timePeriod; float32(len(c.Values)) < max {
		return len(c.Values)
	} else {
		return int(max)
	}
}

func (c channelXYer) XY(index int) (x, y float64) {
	if max := c.freq * timePeriod; float32(len(c.Values)) < max {
		return float64(index), c.Values[index]
	} else {
		return float64((float32(index) - max) / c.freq), c.Values[len(c.Values)-int(max)+index]
	}
}