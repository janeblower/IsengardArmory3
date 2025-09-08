DATABASE<% print(index) %> = {
    <%
    print(
        _.map(sliced, function(account) {
            return `    ["${account.login}"] = {`+ _.map(account.persons, function(person) {
                return `{"${person.name}", ${person.lvl}, ${person.gs}, ${person.race}, "${person.guild}", ${person.class}}`
            }).join(',') +'}'
        })
    
        .join(`,\r\n`))
    %>
    }