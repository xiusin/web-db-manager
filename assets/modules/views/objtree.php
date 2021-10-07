<ul id="tablelist" class="filetree">
<%= if(len(tables) > 0) { %>
    <li id="tables"><span class="tablef"><%= T("Tables") %></span><span class="count"><%= len(tables) %></span>
		<%= for (key, v) in tables { %>
		    <ul><li><span class="file otable" id="t_<%= v.Name %>"><a href='javascript:objDefault("table", "t_<%= v.Name %>")'><%= v.Name %></a></span> <!--不要换行 <span class="count"></span> -->
		    </li>
		    </ul>
		<% } %>
	</li>
<% } %>

<%= if(len(views) > 0) { %>
    <li id="views"><span class="viewf"><%= T("Views") %></span><span class="count"><%= len(views) %></span>
		<%= for (key, v) in views { %>
		    <ul>
				<li><span class="file oview" id="t_<%= v.Name %>"><a href='javascript:objDefault("view", "t_<%= v.Name %>")'><%= v.Name %></a></span></li>
		    </ul>
		<% } %>
	</li>
<% } %>



<%= if(len(procedures) > 0) { %>
    <li id="procs"><span class="procf"><%= T("Procedures") %></span><span class="count"><%= len(procedures) %></span>
		<%= for (key, v) in procedures { %>
		    <ul>
				<li><span class="file oproc" id="p_<%= v.Name %>"><a href='javascript:objDefault("procedure", "p_<%= v.Name %>")'><%= v.Name %></a></span></li>
		    </ul>
		<% } %>
	</li>
<% } %>

<%= if(len(functions) > 0) { %>
    <li id="funcs"><span class="funcf"><%= T("Functions") %></span>
	<span class="count"><%= len(functions) %></span>
		<%= for (key, v) in functions { %>
		    <ul>
				<li><span class="file ofunc" id="f_<%= v.Name %>"><a href='javascript:objDefault("function", "f_<%= v.Name %>")'><%= v.Name %></a></span></li>
		    </ul>
		<% } %>
	</li>
<% } %>

<%= if(len(triggers) > 0) { %>
    <li id="trigs"><span class="trigf"><%= T("Triggers") %></span>
	<span class="count"><%= len(triggers) %></span>
		<%= for (key, v) in triggers { %>
		    <ul>
				<li><span class="file otrig" id="t_<%= v.TriggerName %>"><a href='javascript:objDefault("trigger", "t_<%= v.TriggerName %>")'><%= v.TriggerName %></a></span></li>
		    </ul>
		<% } %>
	</li>
<% } %>

    <li id="events"><span class="evtf"><%= T("Events") %></span>
	<span class="count"><%= len(events) %></span>
		<%= for (key, v) in events { %>
		    <ul>
				<li><span class="file oevt" id="t_<%= v.Name %>"><a href='javascript:objDefault("event", "t_<%= v.Name %>")'><%= v.Name %></a></span></li>
		    </ul>
		<% } %>
	</li>

</ul>

