(function () {
  'use strict';

  var groups = {
    "system": {
      icon: 'fa-user',
      bg: 'success',
      name: 'System'
    },
    "important": {
      icon: 'envelope',
      bg: 'danger',
      name: 'Important'
    },
    "task": {
      icon: 'bell',
      bg: 'warning',
      name: 'Task'
    }
  };

  function app($scope, $websocket, api) {
    var ws;
    $scope.groups = groups;

    $scope.users = {
      mario: {
        name: "Mario",
        username: "mario",
        token: "31a3d1333ddc2e93075593a131158c9ac8ebebe1ea1bc45fdd17a884c2ad5853",
        messages: []
      },
      luigi: {
        name: "Luigi",
        username: "luigi",
        token: "a6ca78a7420c3354315f18f662ef496b4df04e5b7bd41095794ba3653cf55522",
        messages: []
      },
      yoshi: {
        name: "Yoshi",
        username: "yoshi",
        token: "d1a0cf5c56522fc36ffecc5611b6751a20b5a09eee35e86bfb05a2e7d0398dc4",
        messages: []
      }
    };

    $scope.switch = function (username) {
      $scope.user = $scope.users[username];
      ws && ws.reconnect();
    };

    $scope.reset = function() {
      $scope.message = {};
    };

    $scope.send = function() {
      api.message($scope.message, $scope.message.to).success(function (data) {
        $scope.message.data = "";
      }).error(function (data) {
        console.log("failed to send message.", $scope.message, data);
      });
    };

    $scope.remove = function(msg) {
      var idx = $scope.user.messages.indexOf(msg);
      $scope.user.messages.splice(idx, 1);
    };

    $scope.switch("mario");

    ws = $websocket("ws://localhost:8000/ws");

    ws.onMessage(function (msg) {
      $scope.user.messages.push(JSON.parse(msg.data));
    });

    ws.onOpen(function() {
      ws.send(JSON.stringify($scope.user));
    });
  }

  function api($http, $q) {
    return {
      message: function (msg, to) {
        return $http.post('/message', msg, {params: {to: to}});
      }
    };
  }

  angular
    .module('app', ['ui.bootstrap', 'ngWebSocket'])
    .factory('api', api)
    .controller('AppCtrl', app);

})();
