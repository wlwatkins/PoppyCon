$(function () {

	function sidebarToggleOnClick() {
			$('#sidebar-toggle-button').on('click', function () {
				$('#sidebar').toggleClass('sidebar-toggle');
				$('#page-content-wrapper').toggleClass('page-content-toggle');
			});
		}

	(function init() {
		sidebarToggleOnClick();
	})();

});
