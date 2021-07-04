package rediscas

// 提供以下功能:
// 1. GetProto
// 2. SetProto
// 3. Cas机制

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"google.golang.org/protobuf/proto"
)

var (
	// ErrNotExist 数据不存在
	ErrNotExist = errors.New("not exist")
	// ErrCasNoMatch cas冲突
	ErrCasNoMatch = errors.New("cas no match")
)

// Client Redis存储实现
type Conn struct {
	redis.Conn
}

// Get 获取key对应的Val
func (r *Conn) Get(key string) (string, int, error) {
	// lua脚本执行cas解析
	res, err := redis.Values(r.Do("EVAL", getCasLua, "1", key))
	if err != nil && err != redis.ErrNil {
		return "", 0, fmt.Errorf("redis get:%w", err)
	}

	// 判断返回值合法性
	if len(res) != 3 {
		return "", 0, fmt.Errorf("get cas:%w", err)
	}

	exist, _ := redis.Bool(res[0], nil)
	// 不存在
	if !exist {
		return "", 0, ErrNotExist
	}

	cas, _ := redis.Int(res[1], nil)
	val, _ := redis.String(res[2], nil)

	return val, cas, nil
}

// BatchGet 批查
func (r *Conn) BatchGet(keys []string) (map[string]string, map[string]int, error) {
	// 命令调用
	args := redis.Args{}.AddFlat(keys)
	vals, err := redis.Strings(r.Do("MGET", args...))
	if err != nil {
		return nil, nil, err
	}

	// 构造响应
	msgs, cas := make(map[string]string), make(map[string]int)
	for i, val := range vals {
		key := keys[i]
		// 不存在
		if val == "" {
			continue
		}
		msgs[key] = val[8:]
		c := binary.LittleEndian.Uint64([]byte(val[:8]))
		cas[key] = int(c)
	}

	return msgs, cas, nil
}

// Set 更新key对应的Val
func (r *Conn) Set(key, val string, cas int) error {
	// lua脚本执行set命令
	ret, err := redis.Int(r.Do("EVAL", setCasLua, "1", key, val, cas))
	if err != nil {
		return fmt.Errorf("redis set:%w", err)
	}

	if ret != 0 {
		// Cas冲突
		return ErrCasNoMatch
	}

	return nil
}

// SetWithExpire 带过期时间的Set，expire单位为s
func (r *Conn) SetWithExpire(key, val string, cas, expire int) error {
	// lua脚本执行set命令
	ret, err := redis.Int(r.Do("EVAL", setCasExpireLua, "1", key, val, cas, expire))
	if err != nil {
		return fmt.Errorf("redis set:%w", err)
	}

	if ret != 0 {
		// Cas冲突
		return ErrCasNoMatch
	}

	return nil
}

// Del 删除key对应的Val
func (r *Conn) Del(key string) error {
	reply, err := redis.Int(r.Do("DEL", key))
	if err != nil {
		return fmt.Errorf("redis del:%w", err)
	}

	if reply != 1 {
		return fmt.Errorf("redis del ret:%d", reply)
	}
	return nil
}

// GetProto 根据Key获取Val
func (r *Conn) GetProto(key string, msg proto.Message) (int, error) {
	str, cas, err := r.Get(key)
	if err != nil {
		return 0, err
	}

	err = proto.Unmarshal([]byte(str), msg)
	if err != nil {
		return 0, fmt.Errorf("proto unmashal:%w", err)
	}
	return cas, nil
}

// BatchGetProto 批查
func (r *Conn) BatchGetProto(keys []string, msg proto.Message) (map[string]proto.Message,
	map[string]int, error) {
	// 命令调用
	args := redis.Args{}.AddFlat(keys)
	vals, err := redis.Strings(r.Do("MGET", args...))
	if err != nil {
		return nil, nil, err
	}

	// 响应解析
	msgs, cas := make(map[string]proto.Message), make(map[string]int)
	for i, val := range vals {
		key := keys[i]
		// 不存在
		if val == "" {
			continue
		}
		err = proto.Unmarshal([]byte(val[8:]), msg)
		// 深拷贝
		msgs[key] = proto.Clone(msg)
		if err != nil {
			continue
		}

		c := binary.LittleEndian.Uint64([]byte(val[:8]))
		cas[key] = int(c)
	}

	return msgs, cas, nil
}

// SetProto 根据Key更新Val
func (r *Conn) SetProto(key string, info proto.Message, cas int) error {
	b, err := proto.Marshal(info)
	if err != nil {
		return fmt.Errorf("proto mashal:%w", err)
	}

	return r.Set(key, string(b), cas)
}

// SetProtoExpire 根据Key更新Val带过期时间(expire: 秒为单位)
func (r *Conn) SetProtoExpire(key string, info proto.Message, cas, expire int) error {
	b, err := proto.Marshal(info)
	if err != nil {
		return fmt.Errorf("proto mashal:%w", err)
	}

	return r.SetWithExpire(key, string(b), cas, expire)
}
