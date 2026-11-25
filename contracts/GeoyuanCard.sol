// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/// @title GeoyuanCard — 地理元 ERC-721 + 托管市场
/// @notice 每张卡绑定一个唯一的 H3 hexagonal grid cell (uint64)
/// @dev   Polygon PoS / Amoy testnet 兼容
contract GeoyuanCard {

    // ─── 状态 ─────────────────────────────────────────────
    string  public constant name     = "GeoyuanCard";
    string  public constant symbol   = "GYC";
    uint256 public          totalSupply;

    // tokenId → 所有者
    mapping(uint256 => address)                       public ownerOf;
    // tokenId → H3 cell index (uint64)
    mapping(uint256 => uint64)                        public cellOf;
    // H3 cell → tokenId（反查，保证每个 cell 只能 mint 一次）
    mapping(uint64  => uint256)                        public cellMinted;
    // tokenId → metadata URI (or JSON blob 直接存链上)
    mapping(uint256 => string)                        public tokenURI;
    // tokenId → metadata 结构体（也存链上，便宜且省一步）
    mapping(uint256 => CardMeta)                      public metadata;
    // 所有者 → tokenId[]（查 MyCards 用）
    mapping(address => mapping(uint256 => uint256))   public tokensOfOwner;
    mapping(address => uint256)                       public balanceOf;

    // ─── 市场（新增）───────────────────────────────────────
    struct Listing {
        address seller;       // 挂售者
        uint256 price;       // 售价 wei（0 = 免费赠送，仅 gas 成本）
        uint40  listedBlock; // 挂售区块号
    }
    // tokenId → 挂售状态（空 Listing.seller=address(0) 表示未挂售）
    mapping(uint256 => Listing) public listings;
    // 全局挂售 tokenId 列表（用于 marketListings 遍历）
    uint256[] public listedTokenIds;
    // tokenId → 是否在 listedTokenIds 里（防重复）
    mapping(uint256 => bool) public isListedInGlobal;

    struct CardMeta {
        string title;
        string description;
        string color;       // hex e.g. "#c77dff"
        uint64 h3Index;    // H3 cell
        int32 latE4;      // 纬度×10000 (30.2741°N → 302741)
        int32  lonE4;      // 经度×10000 (120.1551°E → 1201551)
        uint32 res;        // H3 resolution
        uint40 timestamp;  // mint 时间戳 (秒)
    }

    // ─── 事件 ─────────────────────────────────────────────
    event Transfer(address indexed from, address indexed to, uint256 indexed tokenId);
    event Minted(uint256 indexed tokenId, address indexed owner, uint64 indexed h3Index);
    event Listed(uint256 indexed tokenId, address indexed seller, uint256 price);
    event Unlisted(uint256 indexed tokenId, address indexed seller);
    event Sold(uint256 indexed tokenId, address indexed seller, address indexed buyer, uint256 price);

    // ─── Errors ────────────────────────────────────────────
    error CellAlreadyMinted(uint64 h3Index, uint256 existingTokenId);
    error NotOwner(uint256 tokenId, address caller);
    error ZeroAddress();
    error EmptyTitle();
    error NotApproved(uint256 tokenId, address caller);
    error NotListed(uint256 tokenId);
    error NotListedByCaller(uint256 tokenId, address caller);
    error InsufficientPayment(uint256 tokenId, uint256 need, uint256 paid);

    // ─── Mint ──────────────────────────────────────────────
    /// @notice 在给定 H3 cell mint 一张 GeoyuanCard
    /// @dev    同一 H3 cell 只能 mint 一次 — 地理唯一性
    function mint(
        uint64 h3Index,
        string calldata title,
        string calldata description,
        string calldata color,
        int32 latE4,
        int32  lonE4,
        uint32 res
    ) external returns (uint256 tokenId) {
        // 唯一性检查：同一 cell 不能 mint 两次
        if (cellMinted[h3Index] != 0) {
            revert CellAlreadyMinted(h3Index, cellMinted[h3Index]);
        }
        if (bytes(title).length == 0) revert EmptyTitle();
        if (msg.sender == address(0)) revert ZeroAddress();

        unchecked { ++totalSupply; }
        tokenId = totalSupply;

        // 核心状态写入
        ownerOf[tokenId]      = msg.sender;
        cellOf[tokenId]       = h3Index;
        cellMinted[h3Index]   = tokenId;
        balanceOf[msg.sender] += 1;
        tokensOfOwner[msg.sender][balanceOf[msg.sender]] = tokenId;

        // Metadata
        metadata[tokenId] = CardMeta({
            title:       title,
            description: description,
            color:       color,
            h3Index:     h3Index,
            latE4:       latE4,
            lonE4:       lonE4,
            res:         res,
            timestamp:   uint40(block.timestamp)
        });
        tokenURI[tokenId] = _buildTokenURI(tokenId);

        emit Transfer(address(0), msg.sender, tokenId);
        emit Minted(tokenId, msg.sender, h3Index);
    }

    // ─── 内部转移（不做 msg.sender 校验，供市场调用）────────
    function _transfer(address from, address to, uint256 tokenId) internal {
        if (ownerOf[tokenId] != from) revert NotOwner(tokenId, from);
        if (to == address(0)) revert ZeroAddress();

        // 从 from 的 tokensOfOwner 里挪掉 tokenId
        // tokensOfOwner 是 1-based 顺序表，O(n) swap-last 移除
        uint256 nFrom = balanceOf[from];
        if (nFrom == 1) {
            delete tokensOfOwner[from][1];
        } else {
            uint256 found = 0;
            for (uint256 i = 1; i <= nFrom; i++) {
                if (tokensOfOwner[from][i] == tokenId) { found = i; break; }
            }
            if (found != 0) {
                tokensOfOwner[from][found] = tokensOfOwner[from][nFrom];
                delete tokensOfOwner[from][nFrom];
            }
        }
        balanceOf[from] -= 1;

        // 加到 to
        balanceOf[to] += 1;
        tokensOfOwner[to][balanceOf[to]] = tokenId;
        ownerOf[tokenId] = to;

        emit Transfer(from, to, tokenId);
    }

    // ─── Transfer (简化版 — 纯直转) ────────────────────────
    function transferFrom(address from, address to, uint256 tokenId) external {
        if (ownerOf[tokenId] != from) revert NotOwner(tokenId, from);
        if (msg.sender != from) revert NotApproved(tokenId, msg.sender);
        _transfer(from, to, tokenId);
    }

    // ─── 市场：挂售 / 下架 / 购买 ──────────────────────────
    /// @notice 挂售一张 GeoyuanCard，token 转到合约托管
    function list(uint256 tokenId, uint256 price) external {
        if (ownerOf[tokenId] != msg.sender) revert NotOwner(tokenId, msg.sender);
        // 已挂售就先 unlist 再 list（允许改价）
        if (listings[tokenId].seller != address(0)) {
            // 允许 owner 自己改价，不强制先 unlist
        }

        _transfer(msg.sender, address(this), tokenId);

        listings[tokenId] = Listing({
            seller:       msg.sender,
            price:        price,
            listedBlock:  uint40(block.number)
        });

        if (!isListedInGlobal[tokenId]) {
            listedTokenIds.push(tokenId);
            isListedInGlobal[tokenId] = true;
        }

        emit Listed(tokenId, msg.sender, price);
    }

    /// @notice 下架自己挂的卡（token 还回 seller）
    function unlist(uint256 tokenId) external {
        Listing storage l = listings[tokenId];
        if (l.seller == address(0)) revert NotListed(tokenId);
        if (l.seller != msg.sender) revert NotListedByCaller(tokenId, msg.sender);

        _transfer(address(this), l.seller, tokenId);
        delete listings[tokenId];
        _removeFromGlobal(tokenId);

        emit Unlisted(tokenId, l.seller);
    }

    /// @notice 买下一张挂售卡（payable）
    function buy(uint256 tokenId) external payable {
        Listing storage l = listings[tokenId];
        if (l.seller == address(0)) revert NotListed(tokenId);
        if (msg.value < l.price) revert InsufficientPayment(tokenId, l.price, msg.value);

        // 多付的 MATIC 退回 buyer
        if (msg.value > l.price) {
            payable(msg.sender).transfer(msg.value - l.price);
        }
        // 售价转给 seller
        payable(l.seller).transfer(l.price);
        // token 从合约转给 buyer
        _transfer(address(this), msg.sender, tokenId);

        address seller = l.seller;
        delete listings[tokenId];
        _removeFromGlobal(tokenId);

        emit Sold(tokenId, seller, msg.sender, l.price);
    }

    function _removeFromGlobal(uint256 tokenId) internal {
        uint256 n = listedTokenIds.length;
        for (uint256 i = 0; i < n; i++) {
            if (listedTokenIds[i] == tokenId) {
                listedTokenIds[i] = listedTokenIds[n - 1];
                listedTokenIds.pop();
                break;
            }
        }
        delete isListedInGlobal[tokenId];
    }

    // ─── 查询便利函数 ──────────────────────────────────────
    /// @notice 查某个 H3 cell 有没有被 mint
    function isCellMinted(uint64 h3Index) external view returns (bool, uint256) {
        uint256 tid = cellMinted[h3Index];
        return (tid != 0, tid);
    }

    /// @notice 查某地址所有卡片（tokenId 列表）
    function myCards(address who) external view returns (uint256[] memory) {
        uint256 n = balanceOf[who];
        uint256[] memory out = new uint256[](n);
        for (uint256 i = 1; i <= n; i++) {
            out[i - 1] = tokensOfOwner[who][i];
        }
        return out;
    }

    /// @notice 邻居匹配：同 cell 附近 k-ring 内有没有别人的卡
    function neighborMatch(address who, uint64[] calldata neighborCells)
        external view returns (uint256[] memory)
    {
        uint256[] memory out = new uint256[](neighborCells.length);
        uint256 found = 0;
        for (uint256 i = 0; i < neighborCells.length; i++) {
            uint256 tid = cellMinted[neighborCells[i]];
            if (tid != 0 && ownerOf[tid] != who) {
                out[found++] = tid;
            }
        }
        uint256[] memory result = new uint256[](found);
        for (uint256 i = 0; i < found; i++) result[i] = out[i];
        return result;
    }

    /// @notice 查全市场挂售卡（分页）
    function marketListings(uint256 limit, uint256 offset)
        external view
        returns (uint256[] memory tokenIds, uint256[] memory prices, address[] memory sellers)
    {
        uint256 n = listedTokenIds.length;
        if (offset >= n) {
            return (new uint256[](0), new uint256[](0), new address[](0));
        }
        uint256 end = offset + limit;
        if (end > n) end = n;
        uint256 size = end - offset;
        tokenIds = new uint256[](size);
        prices   = new uint256[](size);
        sellers  = new address[](size);
        for (uint256 i = 0; i < size; i++) {
            uint256 tid = listedTokenIds[offset + i];
            tokenIds[i] = tid;
            prices[i]   = listings[tid].price;
            sellers[i]  = listings[tid].seller;
        }
    }

    /// @notice 全市场挂售总数
    function listingCount() external view returns (uint256) {
        return listedTokenIds.length;
    }

    /// @notice 查某张卡是否挂售 + 价格 + seller
    function listingOf(uint256 tokenId) external view returns (address seller, uint256 price, uint40 listedBlock) {
        Listing storage l = listings[tokenId];
        return (l.seller, l.price, l.listedBlock);
    }

    // ─── 内部：tokenURI 拼装 ───────────────────────────────
    function _buildTokenURI(uint256 tokenId) internal view returns (string memory) {
        CardMeta storage m = metadata[tokenId];
        return string(abi.encodePacked(
            '{"name":"Geoyuan #', _itoa(int256(tokenId)),
            '","title":"', m.title,
            '","description":"', m.description,
            '","color":"', m.color,
            '","h3Index":"', _itoa(int256(uint256(m.h3Index))),
            '","lat":', _itoa(int256(m.latE4)),
            ',"lon":', _itoa(int256(m.lonE4)),
            ',"res":',  _itoa(int256(uint256(m.res))),
            ',"timestamp":', _itoa(int256(uint256(m.timestamp))),
            '}'
        ));
    }

    // ─── 小工具 ────────────────────────────────────────────
    function _itoa(int256 n) internal pure returns (string memory) {
        if (n == 0) return "0";
        bool neg = n < 0;
        if (neg) n = -n;
        uint256 i;
        uint256 j;
        bytes memory buf = new bytes(78);
        while (n > 0) { buf[i++] = bytes1(uint8(uint256(n % 10) + 48)); n /= 10; }
        if (neg) buf[i++] = '-';
        bytes memory out = new bytes(i);
        for (j = 0; j < i; j++) out[j] = buf[i - j - 1];
        return string(out);
    }
}
