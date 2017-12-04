// init on document ready
$(document).ready(function () {
    init();
});

// get the follower list from the api
function init() {
    var host = "http://localhost:8000";

    // waterfall
    async.waterfall([
        // check
        function (waterfallCb) {
            $.ajax({
                url: host + "/check"
            })
            .done(function (data) {
                waterfallCb(null);
            })
            .fail(function (err) {
                waterfallCb(err);
            });
        },
        // settings
        function (waterfallCb) {
            $.ajax({
                url: host + "/settings"
            })
            .done(function (settings) {
                waterfallCb(null, settings);
            })
            .fail(function (err) {
                waterfallCb(err);
            });
        },
        function (settings, waterfallCb) {
            // check if we are showing followers
            if (!settings.clientShowFollowers) {
                waterfallCb(null, settings, []);
                return;
            }
            
            // get followers
            $.ajax({
                url: host + "/followers"
            })
            .done(function (followers) {
                waterfallCb(null, settings, followers);
            })
            .fail(function (err) {
                waterfallCb(err);
            });
        },
        // shutdown
        function (settings, followers, waterfallCb) {
            /*$.ajax({
                url: host + "/shutdown"
            })
            .done(function (data) {
                waterfallCb(null, settings, followers);
            })
            .fail(function (err) {
                waterfallCb(err);
            });*/
            waterfallCb(null, settings, followers);
        }
    ], function (err, settings, followers) {
        if (err) {
            console.error(err);
            finish(0);
        } else {
            start(settings, followers);
        }
    });
}

// start showing followers
function start(settings, followers) {
    // number of followers
    var count = followers.length;
    
    // time length of outro in milliseconds
    var time = (count) ? settings.clientTimeTotal : 0;

    // reduce time based on follower count and time per
    // this keeps the outro from having too much time betwen followers
    if ((settings.clientTimePer * count) < settings.clientTimeTotal) {
        time = settings.clientTimePer * count;
    }
    
    // interval between each follower
    var interval = time / count;

    // loop through followers
    for (var i = 0; i < followers.length; i++) {
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

// fade in logo when we're done
function finish(time) {
    setTimeout(function () {
        $(".stream-ending").addClass("fade");
        $(".thanks-message").addClass("fade");
        $(".logo").addClass("fade");
    }, time);    
}

// show a follower
function showFollower(follower, i) {
    (function (follower, i) {
        var top, left;
        var user = $("<div>").addClass("user pop-in-out").attr("id", "user-"+i);
        var username = $("<h1>").addClass("username").html(follower.display_name);
        var action = $("<h2>").addClass("action followed").html("followed");
        
        //var times = $("<div>").addClass("times").html("x36");
        //action.append(times);
        
        user.append(username);
        user.append(action);

        var width = getWidth(user);

        var loop = true;
        var attempt = 1;
        while (loop) {
            top = randomTop();
            left = randomLeft(width);
            user.css("top", top+"px").css("left", left+"px")

            if (!hasCollision(width, top, left) || attempt >= 10) {
                loop = false;
                attempt = 1;
            } else {
                attempt++;
            }
        }

        // follower to dom
        $("body").append(user);    

        // set timer for fade out class  (fixes animation retart)
        setTimeout(function () {
            $("#user-" + i).addClass("fade");
        }, 2000);
        
        // set timer for removal from dom (for collision detection and cleanup)
        setTimeout(function () {
            $("#user-"+i).remove();
        }, 3500);
    })(follower, i);
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
    var curShowing = $(".user:not(.fade)");
    
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