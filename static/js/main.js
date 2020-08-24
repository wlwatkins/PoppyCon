$(function () {

  $('.dropdown-trigger').dropdown();
  $('.modal').modal();
  $('.collapsible').collapsible();
  $('select').formSelect();

  $( ".pump" ).click(function(e) {
    console.log(e)
    var button = $(e.currentTarget)
    var org = e.target.firstChild

    console.log(org.data)
    $.ajax({
      url: '/pump',
      type: 'POST',
      dataType: 'json',
      data: {"button": org.data.replace(/\s+/g, '').toLowerCase()},
      beforeSend: function() {
          // setting a timeout
          $(".blockPump").prop("disabled",true);

          button.html( `<span class="spinner-grow spinner-grow-sm" role="status" aria-hidden="true"></span><span class="CountDown"> Loading...</span>` );

      },
      success : function(rawData) {
          counter = 10

        var waterPlants = setInterval(function() {
              counter--;
              // Display 'counter' wherever you want to display it.
              button.html( `<span class="spinner-grow spinner-grow-sm" role="status" aria-hidden="true"></span><span class="CountDown"> ${counter}</span>` );

              if (counter == 0) {
                  // Display a login box
                  button.html( org );
                  $(".blockPump").prop("disabled",false);

                  clearInterval(waterPlants);
              }
          }, 1000);



      },
      error: function(e){
        console.log("error", e);
        button.html( `Error` );

      }
    });

  });

  $( "#sensorSelect" ).change(function(e) {
    $( "#calibSwitch" ).prop( "checked", false );
  });



  $( "#calibSwitch" ).change(function(e) {
    var sensor = $( "#sensorSelect option:selected" ).text();
    console.log($( "#calibSwitch" )[0].checked)

    var mainLoop = setInterval(function(){
      $("#cycleBar").css("width", "100%");
      if ($( "#calibSwitch" )[0].checked && sensor != "Choose sensor") {

        console.log("Ajax call")




          $.ajax({
            url: '/calibration',
            type: 'POST',
            dataType: 'json',
            data: {"sensor": sensor},
            success : function(rawData) {
              console.log(rawData);
              var rawValue = rawData["Data"];
              $("#rawValue").html(rawValue.toFixed(2));

            },
            error: function(e){
              console.log("error", e);
              button.html( `Error` );
            }
          });


        }
        else {
            clearInterval(mainLoop);
            $("#rawValue").html("--");
        }

        $("#cycleBar").css("width", "0%");
    }, 300);


  });





});
