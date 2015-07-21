$(document).ready(function(){
	callback = function(data, status){
  			$("#statuscode").text(status)
  			$("#response").text(data)
        $("#statuscode").show(500)
        $("#response").show(500)
  		};
  var method, data, url
  $("#submitButton").click(function(){
  	$("#statuscode").hide()
  	$("#response").hide()
  	if ($("#postBody").val() == "") {
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
  	} else {
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
  	};

  	
  });

});