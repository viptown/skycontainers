## 모바일
로그인 기능 구현 
로그인후 바로 위치스캔 페이지로 이동

### 툴바는 BOOTOM NAVIGATION으로 구성

1. 위치스캔
    (1): BL 포지션 SELECT 
    (2): 바코드 INPUT 기능 설명: 항상 핸드폰 카메라로 바코드(hblno) 스캔하는 화면이 뜨도록 구현,
    (3): 바코드(hblno) 스캔 후 해당 바코드 값이 bl_markings테이블의 hblno에 있는지 확인, 없으면 없는 비엘이라고 화면에 push해주거 노출시킨다.
    (4): 저장버튼 설명: 클릭시 확인된 바코드(HBLNO)로 BL_MARKINGS의 hblno하고 매칭후 BL_MARKINGS의 bl_position_id 업데이트 한다.
    
2. 비엘위치
    설명: 바코드(HBL) INPUT 기능 설명: INPUT TYPE옆에 핸드폰스캔 버튼 배치 핸드폰 스캔 버튼 클릭 하면  핸드폰 카메마로 바코드(HBLNO) 스캔 기능으로 INPUT 값을 채운다.
    스캔하지 않고 수동으로도 입력 가능

    비엘이 입력된후 검색버튼 클릭하면 
    (1) 해당 비엘이 속한 업체명 (bl_markings테이블의 hblno하고 매칭후 container_id를 통해서 containers테이블에서 업체명을 가져온다.) 
    (2) bl_markings의 bl_position_id에 해당하는 bl_positions의 name필드 값이  업체명과 함께 큰 글자로 화면에 노출 시킨다.
    
3. 휴가신청서
    기존 관리자페이지에 구현되여 있는 축소판으로 구현

구현후 모바일 접속 http주소 제공 답변은 한글로 해줘 