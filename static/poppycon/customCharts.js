$(function () {

	var allCharts = [];




	function renderAllCharts() {
		for (var i = 0; i < allCharts.length; i++)
			allCharts[i].render();
	}

	function toggleDataSeries(e){
		if (typeof(e.dataSeries.visible) === "undefined" || e.dataSeries.visible) {
			e.dataSeries.visible = false;
		}
		else{
			e.dataSeries.visible = true;
		}
		renderAllCharts();
	}


	function convertChartDataToCSV(args) {
	  var result, ctr, keys, columnDelimiter, lineDelimiter, data;

	  data = args.data || null;
	  if (data == null || !data.length) {
	    return null;
	  }

	  columnDelimiter = args.columnDelimiter || ',';
	  lineDelimiter = args.lineDelimiter || '\n';

	  keys = Object.keys(data[0]);

	  result = '';
	  result += keys.join(columnDelimiter);
	  result += lineDelimiter;

	  data.forEach(function(item) {
	    ctr = 0;
	    keys.forEach(function(key) {
	      if (ctr > 0) result += columnDelimiter;
	      result += item[key];
	      ctr++;
	    });
	    result += lineDelimiter;
	  });
	  return result;
	}

	function downloadCSV(args) {
	  var data, filename, link;
	  var csv = "";
	  for(var i = 0; i < args.chart.options.data.length; i++){
	    csv += convertChartDataToCSV({
	      data: args.chart.options.data[i].dataPoints
	    });
	  }
	  if (csv == null) return;

	  filename = args.filename || 'chart-data.csv';

	  if (!csv.match(/^data:text\/csv/i)) {
	    csv = 'data:text/csv;charset=utf-8,' + csv;
	  }

	  data = encodeURI(csv);
	  link = document.createElement('a');
	  link.setAttribute('href', data);
	  link.setAttribute('download', filename);
	  document.body.appendChild(link); // Required for FF
		link.click();
		document.body.removeChild(link);
	}


	function map(x, in_min, in_max, out_min, out_max) {
		return (x-in_min)*(out_max-out_min)/(in_max-in_min) + out_min
	}


	$( document ).ready(function() {
	$.ajax({
		url: '',
		type: 'POST',
		dataType: 'json',
		beforeSend: function() {
				// setting a timeout
				console.log($( ".charts" ))

				$( ".charts" ).append( `<div class="spinner-border" role="status">
				  <span class="sr-only">Loading...</span>
				</div>` );

				setTimeout(function(){
				  $('#someid').addClass("done");
				}, 200000);


		},
		success : function(rawData) {
			console.log(rawData)
			param = rawData['Param'].split("\r\n");
			var param2 = {};
			param2 = $.map(param, function(item) {
				if(item.length>1 && item[0]!= "#"){
				spl = item.split(',')
				param2[spl.slice(0,1)] =  Date.parse(spl.slice(1,2));
				return param2
			}
			});
			param = param2[0]
			console.log(param)

			data = rawData['Data']
			var dat = data["MOISTURE"]
			var dataSeries = []
			var valueLast = 0
			var i = 0
			for (const property in dat) {

				for (const prob in dat[property]) {


				arr = $.map(dat[property][prob]['Data'], function(v, k) {

					var date = new Date(0);
					date.setUTCSeconds(v['Date']);
					var value = map(v['Value'], dat[property][prob]['Calib']["ZeroPCT"], dat[property][prob]['Calib']["HundredPCT"], 0, 100)
					return [{x: date, y: value}]
				});
				i = i+1
				valueLast = valueLast + arr[arr.length - 1]["y"]

				dataSeries = $.merge(dataSeries, [{
						// color: "#393f63",
						markerSize: 0,
						type: "line",
						name: prob,
						showInLegend: true,
						dataPoints: arr
					}]);

			}}
			valueAvg = valueLast/i;

			$( "#humidityTag" ).append("Average: " +valueAvg.toFixed(2)+"%" );

			// CanvasJS spline area chart to show revenue from Jan 2015 - Dec 2015
			var humidityLineChart = new CanvasJS.Chart("humidityLineChart", {
				animationEnabled: true,
				zoomEnabled: true,
			  zoomType: "xy",
				exportEnabled: true,
				backgroundColor: "transparent",
				axisX: {
					// interval: 10,
					// intervalType: "month",
					labelFontColor: "#717171",
					labelFontSize: 16,
					lineColor: "#a2a2a2",
					minimum: new Date(param["startTime"]),
					tickColor: "#a2a2a2",
					valueFormatString: "HH:MM\n D/MMM/YYYY"
				},
				axisY: {
					gridThickness: 0,
					includeZero: false,
					labelFontColor: "#717171",
					labelFontSize: 16,
					lineColor: "#a2a2a2",
					suffix: "%",
					tickColor: "#a2a2a2",
					valueFormatString: "###.## ",
					// minimum: -20,
					// maximum: 120,
				},
				toolTip: {
					borderThickness: 0,
					cornerRadius: 0,
					fontStyle: "normal",
					shared: true,
	        contentFormatter: function(e){
	          var str = "";
						var date = moment(e.entries[0].dataPoint.x).format('HH:mm:ss - DD/MM/YYYY');
						var temp = `${date}<br>`
	          for (var i = 0; i < e.entries.length; i++){

							var serie = `<span style="color:${e.entries[i].dataSeries._colorSet[0]};">${e.entries[i].dataSeries.name}:</span> <strong>${e.entries[i].dataPoint.y.toFixed(2)} %</strong> <br/>`
	            // var  temp = "<span style='"+e.entries[i].dataSeries.name + " <strong>"+  e.entries[i].dataPoint.y + "</strong> <br/>" ;
	            str = str.concat(serie);
							console.log(e);
	          }
						temp = temp.concat(str);
	          return (temp);
	        }
				},
				legend:{
					cursor: "pointer",
					//- fontSize: 16,
					horizontalAlign: "center", // "center" , "right"
					verticalAlign: "top",  // "top" , "bottom"
					itemclick: toggleDataSeries
				},
				data: dataSeries
			});

			humidityLineChart.render();










			var dat = data["TEMPERATURE"]
			var dataSeries = []
			var valueLast = 0
			var i = 0
			for (const property in dat) {

				for (const prob in dat[property]) {

				arr = $.map(dat[property][prob]['Data'], function(v, k) {

					var date = new Date(0);
					date.setUTCSeconds(v['Date']);
					var value = v['Voltage']
					return [{x: date, y: value}]
				});
				i = i+1
				valueLast = valueLast + arr[arr.length - 1]["y"]

				dataSeries = $.merge(dataSeries, [{
						// color: "#393f63",
						markerSize: 0,
						type: "line",
						name: prob,
						showInLegend: true,
						dataPoints: arr
					}]);

			}}

			valueAvg = valueLast/i;

			$( "#temperatureTag" ).append("Average: " +valueAvg.toFixed(1)+"°C" );

			// CanvasJS spline area chart to show revenue from Jan 2015 - Dec 2015
			var temperatureLineChart = new CanvasJS.Chart("temperatureLineChart", {
				animationEnabled: true,
				zoomEnabled: true,
				zoomType: "xy",
				exportEnabled: true,
				backgroundColor: "transparent",
				axisX: {
					// interval: 10,
					// intervalType: "month",
					labelFontColor: "#717171",
					labelFontSize: 16,
					lineColor: "#a2a2a2",
					minimum: new Date(param["startTime"]),
					tickColor: "#a2a2a2",
					valueFormatString: "HH:MM\n D/MMM/YYYY"
				},
				axisY: {
					gridThickness: 0,
					includeZero: false,
					labelFontColor: "#717171",
					labelFontSize: 16,
					lineColor: "#a2a2a2",
					suffix: "°C",
					tickColor: "#a2a2a2",
					valueFormatString: "###.## ",
					// minimum: -5,
					// maximum: 40,
				},
				toolTip: {
					borderThickness: 0,
					cornerRadius: 0,
					fontStyle: "normal",
					shared: true,
	        contentFormatter: function(e){
	          var str = "";
						var date = moment(e.entries[0].dataPoint.x).format('HH:mm:ss - DD/MM/YYYY');
						var temp = `${date}<br>`
	          for (var i = 0; i < e.entries.length; i++){

							var serie = `<span style="color:${e.entries[i].dataSeries._colorSet[0]};">${e.entries[i].dataSeries.name}:</span> <strong>${e.entries[i].dataPoint.y.toFixed(2)}°C</strong> <br/>`
	            // var  temp = "<span style='"+e.entries[i].dataSeries.name + " <strong>"+  e.entries[i].dataPoint.y + "</strong> <br/>" ;
	            str = str.concat(serie);
							console.log(e);
	          }
						temp = temp.concat(str);
	          return (temp);
	        }
				},
				legend:{
					cursor: "pointer",
					//- fontSize: 16,
					horizontalAlign: "center", // "center" , "right"
					verticalAlign: "top",  // "top" , "bottom"
					itemclick: toggleDataSeries
				},
				data: dataSeries
			});

			temperatureLineChart.render();




			var dat = data["LIGHT"]
			var dataSeries = []
			var valueLast = 0
			var i = 0
			for (const property in dat) {

				for (const prob in dat[property]) {


				arr = $.map(dat[property][prob]['Data'], function(v, k) {

					var date = new Date(0);
					date.setUTCSeconds(v['Date']);
					var value = map(v['Value'], dat[property][prob]['Calib']["ZeroPCT"], dat[property][prob]['Calib']["HundredPCT"], 0, 100)
					return [{x: date, y: value}]
				});
				i = i+1
				valueLast = valueLast + arr[arr.length - 1]["y"]

				dataSeries = $.merge(dataSeries, [{
						// color: "#393f63",
						markerSize: 0,
						type: "line",
						name: prob,
						showInLegend: true,
						dataPoints: arr
					}]);

			}}
			valueAvg = valueLast/i;

			$( "#lightTag" ).append("Average: " +valueAvg.toFixed(1)+"%" );

			// CanvasJS spline area chart to show revenue from Jan 2015 - Dec 2015
			var lightLineChart = new CanvasJS.Chart("lightLineChart", {
				animationEnabled: true,
				zoomEnabled: true,
				zoomType: "xy",
				exportEnabled: true,
				backgroundColor: "transparent",
				axisX: {
					// interval: 10,
					// intervalType: "month",
					labelFontColor: "#717171",
					labelFontSize: 16,
					lineColor: "#a2a2a2",
					minimum: new Date(param["startTime"]),
					tickColor: "#a2a2a2",
					valueFormatString: "HH:MM\n D/MMM/YYYY"
				},
				axisY: {
					gridThickness: 0,
					includeZero: false,
					labelFontColor: "#717171",
					labelFontSize: 16,
					lineColor: "#a2a2a2",
					suffix: "%",
					tickColor: "#a2a2a2",
					valueFormatString: "###.## ",
					// minimum: -5,
					// maximum: 40,
				},
				toolTip: {
					borderThickness: 0,
					cornerRadius: 0,
					fontStyle: "normal",
					shared: true,
	        contentFormatter: function(e){
	          var str = "";
						var date = moment(e.entries[0].dataPoint.x).format('HH:mm:ss - DD/MM/YYYY');
						var temp = `${date}<br>`
	          for (var i = 0; i < e.entries.length; i++){
							var serie = `<span style="color:${e.entries[i].dataSeries._colorSet[0]};">${e.entries[i].dataSeries.name}:</span> <strong>${e.entries[i].dataPoint.y.toFixed(2)} %</strong> <br/>`
	            str = str.concat(serie);
	          }
						temp = temp.concat(str);
	          return (temp);
	        }
				},
				legend:{
					cursor: "pointer",
					//- fontSize: 16,
					horizontalAlign: "center", // "center" , "right"
					verticalAlign: "top",  // "top" , "bottom"
					itemclick: toggleDataSeries
				},
				data: dataSeries
			});

			lightLineChart.render();


			allCharts = [humidityLineChart, temperatureLineChart, lightLineChart]

			for (var i = 0; i < allCharts.length; i++){

			var toolBar = document.getElementsByClassName("canvasjs-chart-toolbar")[i];
			var parentChart = toolBar.parentElement.parentElement.id;
			chart = allCharts[i]
			if(chart.get("exportEnabled")){
					var exportCSV = document.createElement('div');
			    var text = document.createTextNode("Save as CSV");
			    exportCSV.setAttribute("style", "padding: 12px 8px; background-color: white; color: black")
			    exportCSV.appendChild(text);

			    exportCSV.addEventListener("mouseover", function(){
			     	 exportCSV.setAttribute("style", "padding: 12px 8px; background-color: #2196F3; color: white")
			    });
			    exportCSV.addEventListener("mouseout", function(){
			      	exportCSV.setAttribute("style", "padding: 12px 8px; background-color: white; color: black")
			    });
			    exportCSV.addEventListener("click", function(){

			      	downloadCSV({ filename: "chart-data.csv", chart: chart })
			    });
					toolBar.lastChild.appendChild(exportCSV);
			  }
			else {
				var exportCSV = document.createElement('button');
				var text = document.createTextNode("Save as CSV");
			  exportCSV.appendChild(text);
			  exportCSV.addEventListener("click", function(){

			      	downloadCSV({ filename: "chart-data.csv", chart: chart })
			  });
			  document.body.appendChild(exportCSV)
			}

}

		},
		error: function(e){
			console.log(e);
		}
	});




});


});
