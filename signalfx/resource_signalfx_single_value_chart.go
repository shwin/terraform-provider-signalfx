package signalfx

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/hashicorp/terraform/helper/schema"
)

func singleValueChartResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the chart",
			},
			"program_text": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Signalflow program text for the chart. More info at \"https://developers.signalfx.com/docs/signalflow-overview\"",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the chart (Optional)",
			},
			"unit_prefix": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "(Metric by default) Must be \"Metric\" or \"Binary\"",
			},
			"color_by": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "(Metric by default) Must be \"Metric\", \"Dimension\", or \"Scale\". \"Scale\" maps to Color by Value in the UI",
			},
			"max_delay": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "How long (in seconds) to wait for late datapoints",
				ValidateFunc: validateMaxDelayValue,
			},
			"refresh_interval": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "How often (in seconds) to refresh the values of the list",
			},
			"max_precision": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum precision to for values displayed in the list",
			},
			"is_timestamp_hidden": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "(false by default) Whether to hide the timestamp in the chart",
			},
			"show_spark_line": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "(false by default) Whether to show a trend line below the current value",
				Default:     false,
			},
			"secondary_visualization": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "(false by default) What kind of secondary visualization to show (None, Radial, Linear, Sparkline)",
				ValidateFunc: validateSecondaryVisualization,
			},
			"color_scale": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Single color range including both the color to display for that range and the borders of the range",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"color": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The color to use. Must be either \"gray\", \"blue\", \"navy\", \"orange\", \"yellow\", \"magenta\", \"purple\", \"violet\", \"lilac\", \"green\", \"aquamarine\"",
							ValidateFunc: validateHeatmapChartColor,
						},
						"gt": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     math.MaxFloat32,
							Description: "Indicates the lower threshold non-inclusive value for this range",
						},
						"gte": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     math.MaxFloat32,
							Description: "Indicates the lower threshold inclusive value for this range",
						},
						"lt": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     math.MaxFloat32,
							Description: "Indicates the upper threshold non-inculsive value for this range",
						},
						"lte": &schema.Schema{
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     math.MaxFloat32,
							Description: "Indicates the upper threshold inclusive value for this range",
						},
					},
				},
			},
			"viz_options": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Plot-level customization options, associated with a publish statement",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"label": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "The label used in the publish statement that displays the plot (metric time series data) you want to customize",
						},
						"color": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Color to use",
							ValidateFunc: validatePerSignalColor,
						},
						"value_unit": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateUnitTimeChart,
							Description:  "A unit to attach to this plot. Units support automatic scaling (eg thousands of bytes will be displayed as kilobytes)",
						},
						"value_prefix": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "An arbitrary prefix to display with the value of this plot",
						},
						"value_suffix": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "An arbitrary suffix to display with the value of this plot",
						},
					},
				},
			},
			"synced": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the resource in the provider and SignalFx are identical or not. Used internally for syncing.",
			},
			"last_updated": &schema.Schema{
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Latest timestamp the resource was updated",
			},
			"url": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL of the chart",
			},
		},

		Create: singlevaluechartCreate,
		Read:   singlevaluechartRead,
		Update: singlevaluechartUpdate,
		Delete: singlevaluechartDelete,
	}
}

/*
  Use Resource object to construct json payload in order to create a single value chart
*/
func getPayloadSingleValueChart(d *schema.ResourceData) ([]byte, error) {
	payload := map[string]interface{}{
		"name":        d.Get("name").(string),
		"description": d.Get("description").(string),
		"programText": d.Get("program_text").(string),
	}

	viz := getSingleValueChartOptions(d)
	if vizOptions := getPerSignalVizOptions(d); len(vizOptions) > 0 {
		viz["publishLabelOptions"] = vizOptions
	}
	if len(viz) > 0 {
		payload["options"] = viz
	}

	return json.Marshal(payload)
}

func getSingleValueChartOptions(d *schema.ResourceData) map[string]interface{} {
	viz := make(map[string]interface{})
	viz["type"] = "SingleValue"
	if val, ok := d.GetOk("unit_prefix"); ok {
		viz["unitPrefix"] = val.(string)
	}
	if val, ok := d.GetOk("color_by"); ok {
		if val == "Scale" {
			if colorScaleOptions := getColorScaleOptions(d); len(colorScaleOptions) > 0 {
				viz["colorBy"] = "Scale"
				viz["colorScale2"] = colorScaleOptions
			}
		} else {
			viz["colorBy"] = val.(string)
		}
	}

	programOptions := make(map[string]interface{})
	if val, ok := d.GetOk("max_delay"); ok {
		programOptions["maxDelay"] = val.(int) * 1000
		viz["programOptions"] = programOptions
	}

	if refreshInterval, ok := d.GetOk("refresh_interval"); ok {
		viz["refreshInterval"] = refreshInterval.(int) * 1000
	}
	if maxPrecision, ok := d.GetOk("max_precision"); ok {
		viz["maximumPrecision"] = maxPrecision.(int)
	}
	if val, ok := d.GetOk("secondary_visualization"); ok {
		secondaryVisualization := val.(string)
		if secondaryVisualization != "" {
			viz["secondaryVisualization"] = secondaryVisualization
		}
	}
	viz["timestampHidden"] = d.Get("is_timestamp_hidden").(bool)
	viz["showSparkLine"] = d.Get("show_spark_line").(bool)

	return viz
}

func singlevaluechartCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalfxConfig)
	payload, err := getPayloadSingleValueChart(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}
	url, err := buildURL(config.APIURL, CHART_API_PATH, map[string]string{})
	if err != nil {
		return fmt.Errorf("[DEBUG] SignalFx: Error constructing API URL: %s", err.Error())
	}

	err = resourceCreate(url, config.AuthToken, payload, d)
	if err != nil {
		return err
	}
	// Since things worked, set the URL and move on
	appURL, err := buildAppURL(config.CustomAppURL, CHART_APP_PATH+d.Id())
	if err != nil {
		return err
	}
	d.Set("url", appURL)
	return nil
}

func singlevaluechartRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalfxConfig)
	path := fmt.Sprintf("%s/%s", CHART_API_PATH, d.Id())
	url, err := buildURL(config.APIURL, path, map[string]string{})
	if err != nil {
		return fmt.Errorf("[DEBUG] SignalFx: Error constructing API URL: %s", err.Error())
	}

	return resourceRead(url, config.AuthToken, d)
}

func singlevaluechartUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalfxConfig)
	payload, err := getPayloadSingleValueChart(d)
	if err != nil {
		return fmt.Errorf("Failed creating json payload: %s", err.Error())
	}
	path := fmt.Sprintf("%s/%s", CHART_API_PATH, d.Id())
	url, err := buildURL(config.APIURL, path, map[string]string{})
	if err != nil {
		return fmt.Errorf("[DEBUG] SignalFx: Error constructing API URL: %s", err.Error())
	}

	return resourceUpdate(url, config.AuthToken, payload, d)
}

func singlevaluechartDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*signalfxConfig)
	path := fmt.Sprintf("%s/%s", CHART_API_PATH, d.Id())
	url, err := buildURL(config.APIURL, path, map[string]string{})
	if err != nil {
		return fmt.Errorf("[DEBUG] SignalFx: Error constructing API URL: %s", err.Error())
	}

	return resourceDelete(url, config.AuthToken, d)
}
