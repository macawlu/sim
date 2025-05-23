/* globals app */

app.controller("DownloadsController", function ($scope, $rootScope, apiget) {

  $scope.$DownloadedFiles = [];
  apiget.files().then(function (xhr) {
    $scope.$DownloadedFiles = xhr.data.Children;
  });

  $scope.$expanded = false;
  $scope.section_expanded_toggle = function () {
    $scope.$expanded = !$scope.$expanded;
    var updateDownloaded = function () {
      apiget.files().then(function (xhr) {
        $scope.$DownloadedFiles = xhr.data.Children;
      }).finally(function () {
        $scope.$applyAsync();
      });
    }
    if ($scope.$expanded) {
      if ($scope.intevID) {
        clearInterval($scope.intevID);
      }
      $scope.intevID = setInterval(updateDownloaded, 30000);
      updateDownloaded();
    } else {
      clearInterval($scope.intevID);
    }
  };
});

app.controller("NodeController", function ($scope, $rootScope, $http, $timeout, reqerr) {
  var n = $scope.node;
  $scope.isfile = function () {
    return !n.Children;
  };
  $scope.isdir = function () {
    return !$scope.isfile();
  };

  var pathArray = [n.Name];
  if ($scope.$parent && $scope.$parent.$parent && $scope.$parent.$parent.node) {
    var parentNode = $scope.$parent.$parent.node;
    pathArray.unshift(parentNode.$path);
    n.$depth = parentNode.$depth + 1;
  } else {
    n.$depth = 1;
  }
  var path = (n.$path = pathArray.join("/"));
  n.$closed = $scope.agoHrs(n.Modified) > 24;
  $scope.audioPreview = /\.(mp3|m4a)$/.test(path);
  $scope.imagePreview = /\.(jpe?g|png|gif)$/.test(path);
  $scope.videoPreview = /\.(mp4|mkv|mov)$/.test(path);

  $scope.isdownloading = function (fileName) {
    if ($scope.isfile() && (fileName in $rootScope.DownloadingFiles)) {
      return true
    }
    return false
  }

  $scope.preremove = function () {
    $scope.confirm = true;
    $timeout(function () {
      $scope.confirm = false;
    }, 3000);
  };

  //defaults
  $scope.closed = function () {
    return n.$closed;
  };
  $scope.toggle = function () {
    n.$closed = !n.$closed;
  };
  $scope.icon = function () {
    var c = [];
    if ($scope.isdownloading(n.Name)) {
      c.push("spinner", "loading");
    } else {
      c.push("outline");
      if ($scope.isfile()) {
        if ($scope.audioPreview) c.push("audio");
        else if ($scope.imagePreview) c.push("image");
        else if ($scope.videoPreview || /\.(avi)$/.test(path)) c.push("video");
        c.push("file");
      } else {
        c.push("folder");
        if (!$scope.closed()) c.push("open");
      }
    }
    c.push("icon");
    return c.join(" ");
  };

  $scope.remove = function (node) {
    $scope.deleting = true;
    $http.delete("download/" + encodeURIComponent(node.$path))
      .then(function () {
        node.$Deleted = true;
        $scope.$applyAsync();
      })
      .catch(reqerr)
      .finally(function () {
        $scope.deleting = false;
      });
  };

  $scope.togglePreview = function () {
    $scope.showPreview = !$scope.showPreview;
  };
});
