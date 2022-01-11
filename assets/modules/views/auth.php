<div class="auth">
	<div>
		<label><%= T("User ID") %>:</label><input type="text" name="auth_user" size="30" value=""/>
	</div>
	<div>
		<label><%= T("Password") %>:</label><input type="password" name="auth_pwd" size="30" />
	</div>

	<div>
		<label><%= T("Server") %>:</label><select name="server" id="server">
		<%= for (key, value) in getServerList() { %>
			<option value="<%= key %>"><%= key %></option>
		<% } %>
		<option value="">自定义</option>
		</select>
	</div>
	<div>
		<label style="margin: 0 5px 0 0;"><%= T("Language") %>:</label>
		<select name="lang">
			<option value="zh">中文(简体)</option>
		</select>
	</div>
	<div id="custom-server" style="display:none">
		<label>服务器地址:</label>
		<input type="text" name="server_name" id="server-name" size="30" value="" />
		<select name="server_type" id="server-type">
			<option value='mysql'>MySQL</option>
		</select>
	</div>
	<div>
		<input type="submit" value="<%= T("Login") %>" />
	</div>
</div>
<script language="javascript" type="text/javascript">
	document.dbform.auth_user.focus();
		document.onready = function() {
		$("#server").change(function() {
			if ($(this).val() == "") {
				$("#custom-server").show();
				$("#server-name").focus();
				$("span.website").hide();
			} else {
				$("#custom-server").hide();
				$("span.website").show();
			}
		});
	}
</script>