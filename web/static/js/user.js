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
    var selectedValues = $(this).data('value');
    console.log(selectedValues);
  });
  $("#submit").click(function() { 
    var data = $('.selection').map(function() {
      return {game: $(this).data('game'), pick: $(this).data('value')};
    }).get();
    console.log(data);
  });
});

