Highcharts.setOptions({
    time: {
        timezone: 'Europe/Paris'
    }
});

$(function () {

  function wait(ms){
     var start = new Date().getTime();
     var end = start;
     while(end < start + ms) {
       end = new Date().getTime();
    }
  }

  function map(x, in_min, in_max, out_min, out_max) {
    return (x - in_min) * (out_max - out_min) / (in_max - in_min) + out_min;
  }



  function drawChart(data, htmlID, chartID, chartTitle, chartYAxis, unit) {


    var series = [];
    var avg = 0;
    for (var i = 0; i < data.length; i++) {
        var s = data[i];
        var label = Object.keys(s)[0];
        var obj = s[label];
        var calib0 = obj["Calib"]["ZeroPCT"]
        var calib100 = obj["Calib"]["HundredPCT"]
        var orgData = obj["Data"]
        var procData =[];
        
        var v;
        for (var j = 0; j < orgData.length; j++) {
          console.log(data)
          if (!calib0 == "") {
            v = map(orgData[j]["Value"], calib0, calib100, 0, 100)
          } else {
            v = orgData[j]["Value"]
          }
          procData.push([orgData[j]["Date"]*1000, v]);
        }

        // if (procData.length) {
        //     avg = avg + procData.slice(-1)[0][1]
        // }
        avg = avg + procData.slice(-1)[0][1]
        series.push({"name": label, "data": procData})
        //Do something

    }

    avg = avg / data.length
    $(htmlID).append(avg.toFixed(2), unit)





    Highcharts.chart(chartID, {
        chart: {
            zoomType: 'xy'
        },
        title: {
            text: chartTitle
        },
        tooltip: {
            valueDecimals: 2
        },

        yAxis: {
            title: {
                text: chartYAxis
            }
        },

        xAxis: {
          title: {
              text: 'Date time'
          },
          type: 'datetime',
          plotBands: [{
            color: 'orange', // Color value
            from: 1167696000000, // Start of the plot band
            to: 1167868800000 // End of the plot band
          }],
        },

        legend: {
            layout: 'vertical',
            align: 'right',
            verticalAlign: 'middle'
        },

        plotOptions: {
            series: {
                turboThreshold:5000,
                label: {
                    connectorAllowed: false
                },
                pointStart: 2010
            }
        },
        series: series,
        responsive: {
            rules: [{
                condition: {
                    maxWidth: 500,
                },
                chartOptions: {
                    legend: {
                        layout: 'horizontal',
                        align: 'center',
                        verticalAlign: 'bottom'
                    }
                }
            }]
        }

    });

    return avg

  }

  $( document ).ready(function() {
  	$.ajax({
  		url: '',
  		type: 'POST',
  		dataType: 'json',
  		beforeSend: function() {
          console.log("Before sending ajax")
  		},
  		success : function(rawData) {

  			console.log(rawData);
        console.log("##################");

        // MOISTER STUFF
        var moistureData = rawData["Data"]['MOISTURE'];

        if (moistureData != undefined){

          avg = drawChart(moistureData, "#humPlant", 'chartHum', 'Humidity', 'Moisture [%]', "%");

          if (avg < 20) {
            $("#humPlant").addClass("amber darken-2");
            $($("#humPlant")[0].parentElement.parentElement.parentElement.children[0].firstElementChild).removeClass("hide");
          } else if (avg >= 20 && avg > 80) {
            $("#humPlant").addClass("light-green darken-2");
          } else if (avg >= 80) {
            $("#humPlant").addClass("light-blue darken-3");
            $($("#humPlant")[0].parentElement.parentElement.parentElement.children[0].firstElementChild).removeClass("hide");
          }
        }


        var temperatureData = rawData["Data"]['TEMPERATURE'];

        if (temperatureData != undefined){

          avg = drawChart(temperatureData, "#tempPlant", 'chartTemp', 'Temperature', 'Temperature [°C]', "°C");

          if (avg <= 0) {
            $("#tempPlant").addClass("blue darken-3");
            $($("#tempPlant")[0].parentElement.parentElement.parentElement.children[0].firstElementChild).removeClass("hide");
          } else if (avg > 0 && avg <= 25) {
            $("#tempPlant").addClass("light-green darken-2");
          } else if (avg > 25) {
            $("#tempPlant").addClass("orange darken-4");
            $($("#tempPlant")[0].parentElement.parentElement.parentElement.children[0].firstElementChild).removeClass("hide");
          }

        }

        var tempDHTData = rawData["Data"]['dhtTemp'];

        if (tempDHTData != undefined){

          avg = drawChart(tempDHTData, "#tempAir", 'chartDHTTemp', 'Ambient temperature', 'Temperature [°C]', "°C");

          if (avg <= 0) {
            $("#tempAir").addClass("blue darken-3");
            $($("#tempAir")[0].parentElement.parentElement.parentElement.children[0].firstElementChild).removeClass("hide");
          } else if (avg > 0 && avg <= 25) {
            $("#tempAir").addClass("light-green darken-2");
          } else if (avg > 25) {
            $("#tempAir").addClass("orange darken-4");
            $($("#tempAir")[0].parentElement.parentElement.parentElement.children[0].firstElementChild).removeClass("hide");
          }

        }

        var moistDHTData = rawData["Data"]['dhtHum'];

        if (moistDHTData != undefined){

          avg = drawChart(moistDHTData, "#humAir", 'chartDHTHum', 'Ambient humidity', 'Moisture [%]', "%");

          if (avg < 20) {
            $("#humAir").addClass("amber darken-2");
            $($("#humAir")[0].parentElement.parentElement.parentElement.children[0].firstElementChild).removeClass("hide");
          } else if (avg >= 20 && avg > 80) {
            $("#humAir").addClass("light-green darken-2");
          } else if (avg >= 80) {
            $("#humAir").addClass("light-blue darken-3");
            $($("#humAir")[0].parentElement.parentElement.parentElement.children[0].firstElementChild).removeClass("hide");
          }
        }



        var lightData = rawData["Data"]['LIGHT'];

        if (lightData != undefined){
          avg = drawChart(lightData, "#lightPlant", 'chartLight', 'Light', 'Light [%]', "%");
        }


        var waterData = rawData["Water"]

        if (waterData != undefined){

          var series = [];

          for (var i = 0; i < waterData.length; i++) {


              series.push({"x": waterData[i]["Date"]*1000, "label": waterData[i]['Pump'], "name": waterData[i]['Pump']})

          }




          Highcharts.chart('chartTimeline', {
              chart: {
                  zoomType: 'x',
                  type: 'timeline'
              },
              xAxis: {
                  type: 'datetime',
                  visible: true
              },
              yAxis: {
                  gridLineWidth: 1,
                  title: null,
                  labels: {
                      enabled: false
                  }
              },
              legend: {
                  enabled: false
              },

              plotOptions: {
                  series: {
                      turboThreshold:5000,
                      label: {
                          connectorAllowed: false
                      },
                      pointStart: 2010
                  }
              },
              title: {
                  text: 'Watering timeline'
              },
              tooltip: {
                  // style: {
                  //     width: 300
                  // },
                  enabled: true,

              },
              series: [{
                  dataLabels: {
                      allowOverlap: false,
                      format: '<span style="color:{point.color}">● </span><span style="font-weight: bold;" > ' +
                          '{point.x:%d %b %Y - %H:%M}</span><br/>{point.label}'
                  },
                  marker: {
                      symbol: 'circle'
                  },
                  data: series
              }]
          });

        }











        $( ".progress" ).hide();

        let date = new Date().toLocaleDateString("en", {year:"numeric", day:"2-digit", month:"short"});
        let time = new Date().toLocaleTimeString("fr"); // 11:18:48 AM
        $("#currentDate").append(date)
        $("#currentTime").append(time)




  		},
  		error: function(e){
  			console.log(e);
  		}
  	});
  });







});
