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
  $('.joinLeagueForm').submit(function(e){
    e.preventDefault();
    let league = $('#name').val();

    $.ajax({
        url: '/api/leagues/' + league + '/users',
        type: 'put',
        data:$('.joinLeagueForm').serialize(),
        success:function(){
          window.location.href = "/user"
        }
    });
  });
});

