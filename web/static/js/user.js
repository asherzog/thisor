$(document).ready(function() {
  $('.userCreateForm').submit(function(e){
    e.preventDefault();
    $.ajax({
        url: '/api/users',
        type: 'post',
        data:$('.userCreateForm').serialize(),
        success:function(){
          window.location.href = "/home"
        }
    });
  });
});

