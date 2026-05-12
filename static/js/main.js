

        document.getElementById("submit").addEventListener("click",()=>{
            post()
        })

        function post(){
            let body={"domain":domain.value,"type":type.value, "server":server.value}

            let code=200
            fetch("/resolve/",  {method: 'POST',headers: {'Content-Type': 'application/json'}, body: JSON.stringify(body)})
            .then(resp=>{
                if (resp.ok){
                    return resp.json()
                }
                else{
                    code=resp.status
                    return
                }

            })
            .then(res=>{
                if (code!==200){
                    if (code==429){
                        result.value="Too many requests (429)"
                    } else{
                        result.value=`Unexpected error (${code})`
                    }
                    return
                }
                if (!isIterable(res)){
                    result.value = "DNS record not found."
                    return
                }
                let text =[]
                let i=0
                for (let el of res){
                    
                    text.push(`${i}\tTTL\t\t${el.ttl}\t${el.type}\t${el.value}`)
                    i++
                }

                result.value=text.join("\n\n")
            })

            
        }

        function isIterable(obj) {
        if (obj == null) return false;
        
        return typeof obj[Symbol.iterator] === 'function';
        }


        function setCookie(name, value, options = {}) {

            options = {
                path: '/',
                ...options
            };

            if (options.expires instanceof Date) {
                options.expires = options.expires.toUTCString();
            }

            let updatedCookie = encodeURIComponent(name) + "=" + encodeURIComponent(value);

            for (let optionKey in options) {
                updatedCookie += "; " + optionKey;
                let optionValue = options[optionKey];
                if (optionValue !== true) {
                updatedCookie += "=" + optionValue;
                }
            }

            document.cookie = updatedCookie;
        }
        function deleteCookie(name) {
            setCookie(name, "", {
                'max-age': -1
            })
        }

        
        server.addEventListener("change",()=>{
            setCookie('server', server.value, {secure: false, 'max-age': 259200});
        })
        type.addEventListener("change",()=>{
            setCookie('type', type.value, {secure: false, 'max-age': 259200});
        })
        domain.addEventListener("change",()=>{
            setCookie('domain', domain.value, {secure: false, 'max-age': 259200});
        })
