package rediscas

const (
	// getCasLua 获取数据后解析cas信息
	getCasLua = `
		local key=KEYS[1];
		local raw_content=redis.call('get', key);
		local real_content;
		local raw_cas=0;

		if raw_content == false then
		    return {false, 0, ''};
		else
		    raw_cas = struct.unpack("l", raw_content);
		    real_content = string.sub(raw_content, 9);
		    return {true, raw_cas, real_content};
		end
	`

	// setCas 添加cas后存储信息
	setCasLua = `
		local key=KEYS[1];
		local set_content=ARGV[1];
		local set_cas=tonumber(ARGV[2]);
		local raw_cas=0;
		local raw_content=redis.call('get', key);

		if raw_content == false then
		    raw_cas=0;
		else
		    raw_cas = struct.unpack("l", raw_content);
		end

		if set_cas < 0 or set_cas >= raw_cas then
		    if raw_cas >= 2147483647 then
		        raw_cas = 0;
			end
		    raw_content=struct.pack("lc0", raw_cas+1, set_content);
		    redis.call('set', key, raw_content);
		    return 0;
		else
		    return -1;
		end
	`

	// setCasExpire 添加cas后存储信息
	setCasExpireLua = `
		local key=KEYS[1];
		local set_content=ARGV[1];
		local set_cas=tonumber(ARGV[2]);
		local expire=tonumber(ARGV[3]);
		local raw_cas=0;
		local raw_content=redis.call('get', key);

		if raw_content == false then
		    raw_cas=0;
		else
		    raw_cas = struct.unpack("l", raw_content);
		end

		if set_cas < 0 or set_cas >= raw_cas then
		    if raw_cas >= 2147483647 then
		        raw_cas = 0;
			end
		    raw_content=struct.pack("lc0", raw_cas+1, set_content);
		    redis.call('set', key, raw_content, 'EX', expire);
		    return 0;
		else
		    return -1;
		end
	`
)
