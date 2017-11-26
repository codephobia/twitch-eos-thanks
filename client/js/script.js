// init on document ready
$(document).ready(function () {
    init();
});

// get the follower list from the api
function init() {
    $.ajax({
        url: "http://localhost:8000/followers"
    })
    .done(function (data) {
        start(data);
    })
    .fail(function () {
        finish(0);
    });
}

// start showing followers
function start(followers) {
    var time = .5 * 60 * 1000;
    var count = followers.length;

    // loop through followers
    for (var i = 0; i < followers.length; i++) {
        // interval between each follower
        var interval = time / count;
        
        // time to show this follower
        var t = Math.floor(interval * i);
        
        // set timeout for showing follower
        (function (i, t) {
            setTimeout(function () {
                showFollower(followers[i], i);
            }, t);
        })(i, t);
    }
    
    finish(time);
}

function finish(time) {
// fade in logo when we're done
    setTimeout(function () {
        $(".stream-ending").addClass("fade");
        $(".thanks-message").addClass("fade");
        $(".logo").addClass("fade");
    }, time);    
}

// show a follower
function showFollower(follower, i) {
    var top, left;
    var user = $("<div>").addClass("user").attr("id", "user-"+i);
    var username = $("<h1>").addClass("username").html(follower.UserData.display_name);
    var action = $("<h2>").addClass("action followed").html("FOLLOWED");
    
    user.append(username);
    user.append(action);
    
    var width = getWidth(user);
    
    var loop = true;
    while (loop) {
        top = randomTop();
        left = randomLeft(width);
        user.css("top", top+"px").css("left", left+"px")

        if (!hasCollision(width, top, left)) {
            loop = false;
        }
    }
    
    $("body").append(user);
    
    (function (i) {
        setTimeout(function () {
            $("#user-"+i).addClass("fade");
        }, 1000);
        setTimeout(function () {
            $("#user-"+i).remove();
        }, 3500);
    })(i);
}

// get width of a user element
function getWidth(user) {
    // clone user
    var newUser = user.clone(user, true);
    
    // add css / class
    newUser.addClass("temp-user").css("visibility", "hidden");
    
    // append clone
    $("body").append(newUser);
    
    // get width of clone
    var width = $(".temp-user").width();
    
    // remove clone
    $(".temp-user").remove();
    
    // return width
    return width;
}

// get a random top position
function randomTop() {
    var height = $(window).height()
    var marginTop = 150;
    var marginBottom = 150;
    var userHeight = 135;
    
    var min = marginTop;
    var max = height - (marginBottom + userHeight);
    
    return Math.floor(Math.random() * ((max-min)+1) + min);
}

// get a random left position
function randomLeft(userWidth) {
    var width = $(window).width();
    var min = 10;
    var max = width - userWidth - 30;
    
    return Math.floor(Math.random() * ((max-min)+1) + min);
}

// check for collision with existing users
function hasCollision(width, top, left) {
    // current users showing
    var curShowing = $(".user");
    
    console.log(curShowing)
    
    // new user bounding box
    var rect1 = {
        top: top,
        right: left + width,
        bottom: top + 135,
        left: left
    };
    
    // loop through existing users checking for overlaps
    for (var i = 0; i < curShowing.length; i++) {
        // bounding box of user
        var position = $(curShowing[i]).position();
        var rect2 = {
            top: position.top,
            right: position.left + $(curShowing[i]).width(),
            bottom: position.top + $(curShowing[i]).height(),
            left: position.left
        };
        
        console.log(rect2);
        
        // check for overlap
        var overlap = !(rect1.right < rect2.left || 
                    rect1.left > rect2.right || 
                    rect1.bottom < rect2.top || 
                    rect1.top > rect2.bottom);
        
        // if overlap return collision
        if (overlap) {
            return true;
        }
    }
    
    // if no overlaps return false
    return false;
}