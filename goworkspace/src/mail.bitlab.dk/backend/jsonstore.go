package backend
import "strconv"

const (
	MEMORY_JSON_STORE_UID_KEY = "__uid__";
)

type JSonStore interface {
	GetJSonBlob(matching map[string]string) []map[string]string;
	PutJSonBlob(jsonblob map[string]string) uint64;
}


type MemoryJsonStore struct {
	memory map[uint64]map[string]string;
	idpool uint64;
}

func (ths *MemoryJsonStore) GetJSonBlob(matching map[string]string) []map[string]string {
	var res = make([]map[string]string,0);
	//var ids = make([]uint64,0);
	for id, val := range ths.memory {
		for key,v := range val {
			var allMatch = true;
			for skey,sv := range matching {
				if (key != skey && v != sv) {
					allMatch = false;
					break;
				}
			}
			if (allMatch) {
				res = append(res,ths.memory[id]);
			//	ids = append(ids,id);
			}
		}
	}
	return res;
}

func (ths *MemoryJsonStore) PutJSonBlob(jsonblob map[string]string) uint64 {
	id := ths.idpool+1;
	ths.idpool += 1;
	ths.memory[id] = jsonblob;
	jsonblob[MEMORY_JSON_STORE_UID_KEY] = strconv.FormatUint(id,10);
	return id;
}

func NewMemoryStore() JSonStore {
	var result *MemoryJsonStore = new(MemoryJsonStore);
	result.memory = make(map[uint64]map[string]string);
	result.idpool = 0;
	return result;
}