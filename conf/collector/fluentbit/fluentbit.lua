local reassemble_state = {}
function reassemble_cri_logs(tag, timestamp, record)
  -- IMPORTANT: reassemble_key must be unique for each parser stream
  -- otherwise entries from different sources will get mixed up.
  -- Either make sure that your parser tags satisfy this or construct
  -- reassemble_key some other way
  local reassemble_key = tag
  -- if partial line, accumulate
  if record.logtag == 'P' then
     if reassemble_state[reassemble_key] == nil then
       reassemble_state[reassemble_key] = ""
     end
     if record.message ~= nil then
       reassemble_state[reassemble_key] = reassemble_state[reassemble_key] .. record.message
     end
     return -1, 0, 0
  end
  modCode = 2
  -- otherwise it's a full line, concatenate with accumulated partial lines if any
  if reassemble_state[reassemble_key] == nil then
     modCode = 0
     reassemble_state[reassemble_key] = ""
  end
  record.message = reassemble_state[reassemble_key] .. (record.message or "")
  reassemble_state[reassemble_key] = nil
  return modCode, timestamp, record
end

e_checks = {}
e_checks["debug"] = {"(%Debug)","%(DEBUG)","(%[D]%d)","(%level=debug)","(%Value:debug)","(%\"level\":\"debug\")"}
e_checks["warn"]  = {"(%Warn)","%(WARN)","(%[W]%d)","(%level=warn)","(%Value:warn)","(%\"level\":\"warn\")"}
e_checks["info"]  = {"(%Info)" ,"(%INFO)" ,"(%[I]%d)","(%level=info)","(%Value:info)" ,"(%\"level\":\"info\")"}
e_checks["error"] = {"(%Error)","(%ERROR)","(%[E]%d)","(%level=error)","(%Value:error)","(%\"level\":\"error\")"}
function extract_log_level(message)
   match = nil
   for j,level in ipairs({"debug","warn","info","error"}) do
       for i,check in ipairs(e_checks[level]) do
         match = string.match(message,check)
         if match ~= nil then
             return level
         end
       end
   end
   return nil
end
function transform(tag, timestamp, record)
   ce = {}
   ce['logtag'] = record.logtag
   ce['stream'] = record.stream
   ce['filename'] = record.path
   ce['container_id'] = record.docker_id
   ce['container_image'] = record.container_image
   record["docker"] = ce
   openshift = {}
   openshift['labels'] = {}
   record['openshift'] = openshift
   pipeline = {}
   collector = {}
   collector["name"] = "fluent-bit"
   collector["version"] = "someversion"
   collector["ipaddr4"] = os.getenv("NODE_IPV4")
   collector["received_at"] = os.date("%Y-%m-%dT%H:%M:%S")
   collector["inputname"] = "kubernetes"

   kube = record["kubernetes"]
   kube["container_name"] = record["container_name"]

   pipeline["collector"] = collector

   record["pipeline_metadata"] = pipeline

   level = extract_log_level(record["message"])
   if level ~= nil then
     record["level"] = level
   end
   return 1, timestamp, record
end



