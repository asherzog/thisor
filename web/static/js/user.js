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
  $('.parentDiv').delegate('.childDiv', 'click', function() {
    $(this).addClass('selection').siblings().removeClass('selection');
  });
  $("#submit").click(function() { 
    var user = ""
    var data = $('.selection').map(function() {
      user = $(this).data('user')
      return {
        game_id: $(this).data('game').toString(),
        user_id: user, 
        selection: {id: $(this).data('value').toString()},
        week: $(this).data('week')
      };
    }).get();
    var win = $("#win").val();
    var lose = $("#lose").val();
    var len = data.length
    if (len > 0) {
      data[len-1]["win_score"] = parseInt(win)
      data[len-1]["lose_score"] = parseInt(lose)
    }
    var redirect = $('#backtoleague').attr('href');
    $.ajax({
      type: "POST",
      url: "/api/picks/list",
      data: JSON.stringify({ "users": {user:data} }),
      contentType: "application/json; charset=utf-8",
      dataType: "json",
      success:function(){
        window.location.href = redirect
      }
    });
  });
});

