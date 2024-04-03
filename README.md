# DMX/Artnet Manipulator

## User config

### Иерархия
```
dmx_devices:
  - alias: DMX1
    path: "COM1" 
    scenes:
    - scene_alias: "scene 1"
      channel_map:
      -  scene_channel_id: 0
         universe_channel_id: 11
      -  scene_channel_id: 1
         universe_channel_id: 12
      -  scene_channel_id: 2
         universe_channel_id: 13   
      -  scene_channel_id: 3
         universe_channel_id: 14
      -  scene_channel_id: 4
         universe_channel_id: 15
      -  scene_channel_id: 5
         universe_channel_id: 16
      -  scene_channel_id: 6
         universe_channel_id: 17   
      -  scene_channel_id: 7
         universe_channel_id: 18
    - scene_alias: "scene 2"
      channel_map:
      -  scene_channel_id: 0
         universe_channel_id: 20
      -  scene_channel_id: 1
         universe_channel_id: 25 
artnet_devices:
  - alias: Artnet1
    ip: "172.18.191.119"
    scenes:
    - scene_alias: "scene 1"
      channel_map:
      -  scene_channel_id: 0
         universe_channel_id: 11
      -  scene_channel_id: 1
         universe_channel_id: 12
      -  scene_channel_id: 2
         universe_channel_id: 13   
      -  scene_channel_id: 3
         universe_channel_id: 14
      -  scene_channel_id: 4
         universe_channel_id: 15
      -  scene_channel_id: 5
         universe_channel_id: 16
      -  scene_channel_id: 6
         universe_channel_id: 17   
      -  scene_channel_id: 7
         universe_channel_id: 18
    - scene_alias: "scene 2"
      channel_map:
      -  scene_channel_id: 0
         universe_channel_id: 20
      -  scene_channel_id: 1
         universe_channel_id: 25
  ...
```
### Атрибуты

#### dmx_devices 

Тип аргументов: Array   
   
Описание: В данной секции необходимо перечислить все используемые DMX устройства.

#### artnet_devices 

Тип аргументов: Array   
   
Описание: В данной секции необходимо перечислить все используемые Artnet устройства.

#### alias 

Тип аргументов: String  
   
Описание: Пользовательский идентификатор устройства.

Ограничения: Должен быть уникальным для всех DMX и Artnet устройств одновременно.

#### path (DMX)

Тип аргументов: String   
   
Описание: В данной секции необходимо указать название порта DMX устройства, как он определен в системе.

Список доступных портов DMX устройств в системе можно узнать через список всех USB устройств в системе
```
ls /dev
```
#### ip (Artnet)

Тип аргументов: String   
   
Описание: В данной секции необходимо указать IP-адрес Artnet контроллера.

#### scenes 

Тип аргументов: Array   
   
Описание: Набор сцен для конкретного устройства. Список может быть пустым.

#### scene_alias 

Тип аргументов: String  
   
Описание: Пользовательский идентификатор сцены.

Ограничения: Должен быть уникальным для списка сцен конкретного устройства, иначе будет переопределен во время сереализации последней записью с повторяющимся названием. 

#### channel_map 

Тип аргументов: Array   
   
Описание: Список каналов, используемых в конкретной сцене, связывающий относительные и абсолютные индексы каналов в universe.

#### scene_channel_id 

Тип аргументов: Integer   
   
Описание: Относительный индекс канала в конкретной сцене.

Ограничения: Должен быть уникальным для списка каналов конкретной сцены, иначе будет переопределен во время сереализации последней записью с повторяющимся индексом.

#### universe_channel_id 

Тип аргументов: Integer   
   
Описание: Абсолютный индекс канала в universe.

Ограничения: Должен совпадать с диапазоном используемых каналов в DMX/Artnet [1;512].
