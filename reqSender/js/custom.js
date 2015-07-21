$(document).ready(function(){
	callback = function(data, status,xhr){
  			$("#statuscode").text(xhr.status)
  			$("#response").text(data)
        $("#statuscode").show(500)
        $("#response").show(500)
  		};
  var method, data, url

  $("#get").click(function(){
  	$("#statuscode").hide()
  	$("#response").hide()

		method = "GET"
		$.ajax({
          type:method,
          beforeSend: function (request)
          {
              request.setRequestHeader("Authorization", "QiniuStub uid=1&ut=4");
          },
          url: $("#url").val() + $("#uri").val() ,
          processData: false,
          success: callback
    });
  });

  $("#post").click(function() {
      $("#statuscode").hide()
      $("#response").hide()
  		method = "POST"
  		data = $("#postBody").val()
  		$.ajax({
            type:method,
            beforeSend: function (request)
            {
                request.setRequestHeader("Authorization", "QiniuStub uid=1&ut=4");
            },
            url: $("#url").val() + $("#uri").val(),
            data: data,
            processData: false,
            success: callback
      });
  	});

  	
  });
